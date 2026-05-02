import { spawn } from 'node:child_process';
import { createHash } from 'node:crypto';
import { existsSync, mkdirSync, readFileSync, writeFileSync, chmodSync } from 'node:fs';
import { get } from 'node:https';
import { homedir } from 'node:os';
import { basename, join } from 'node:path';
import { Readable } from 'node:stream';

const REPO_OWNER = 'ricardoborges-teachable';
const REPO_NAME = 'ai-setup';
const LATEST_RELEASE_API_URL = `https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest`;
const LATEST_DOWNLOAD_BASE_URL = `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download`;
const MAX_REDIRECTS = 5;

interface LatestRelease {
  tag_name?: string;
}

interface DownloadResult {
  buffer: Buffer;
  finalUrl: string;
}

export function bootstrap(): void {
  void run().catch((error: unknown) => {
    console.error(`ai-setup: ${formatError(error)}`);
    process.exit(1);
  });
}

async function run(): Promise<void> {
  const binaryName = getBinaryName();
  const cacheDir = getCacheDir();
  const binaryPath = join(cacheDir, binaryName);
  const versionPath = join(cacheDir, '.version');

  mkdirSync(cacheDir, { recursive: true });

  const hasCachedBinary = existsSync(binaryPath);
  const cachedVersion = readCachedVersion(versionPath);

  let binaryReady = hasCachedBinary;

  try {
    const latestVersion = await getLatestReleaseVersion();

    if (!hasCachedBinary || cachedVersion !== latestVersion) {
      await downloadAndInstallBinary({ binaryName, binaryPath, versionPath, version: latestVersion });
      binaryReady = true;
    }
  } catch (error) {
    if (!hasCachedBinary) {
      throw error;
    }

    console.warn(`ai-setup: unable to check for updates, using cached binary${cachedVersion ? ` (${cachedVersion})` : ''}: ${formatError(error)}`);
    binaryReady = true;
  }

  if (!binaryReady) {
    throw new Error('no ai-setup binary is available');
  }

  await execCachedBinary(binaryPath);
}

function getBinaryName(): string {
  const osName = mapPlatform(process.platform);
  const archName = mapArch(process.arch);
  const extension = osName === 'windows' ? '.exe' : '';

  return `ai-setup-${osName}-${archName}${extension}`;
}

function mapPlatform(platform: NodeJS.Platform): 'darwin' | 'linux' | 'windows' {
  if (platform === 'darwin') {
    return 'darwin';
  }

  if (platform === 'linux') {
    return 'linux';
  }

  if (platform === 'win32') {
    return 'windows';
  }

  throw new Error(`unsupported platform: ${platform}`);
}

function mapArch(arch: NodeJS.Architecture): 'arm64' | 'amd64' {
  if (arch === 'arm64') {
    return 'arm64';
  }

  if (arch === 'x64') {
    return 'amd64';
  }

  throw new Error(`unsupported architecture: ${arch}`);
}

function getCacheDir(): string {
  const home = homedir() || process.env.USERPROFILE;

  if (!home) {
    throw new Error('unable to determine home directory');
  }

  return join(home, '.ai-setup', 'bin');
}

function readCachedVersion(versionPath: string): string | undefined {
  if (!existsSync(versionPath)) {
    return undefined;
  }

  const version = readFileSync(versionPath, 'utf8').trim();
  return version.length > 0 ? version : undefined;
}

async function getLatestReleaseVersion(): Promise<string> {
  const release = await getJson<LatestRelease>(LATEST_RELEASE_API_URL);

  if (!release.tag_name) {
    throw new Error('latest GitHub release response did not include tag_name');
  }

  return release.tag_name;
}

async function downloadAndInstallBinary(input: {
  binaryName: string;
  binaryPath: string;
  versionPath: string;
  version: string;
}): Promise<void> {
  const binaryUrl = `${LATEST_DOWNLOAD_BASE_URL}/${input.binaryName}`;
  const checksumsUrl = `${LATEST_DOWNLOAD_BASE_URL}/checksums.txt`;

  const [binaryDownload, checksumsDownload] = await Promise.all([
    download(binaryUrl),
    download(checksumsUrl),
  ]);

  const expectedChecksum = parseChecksum(checksumsDownload.buffer.toString('utf8'), input.binaryName);
  const actualChecksum = sha256(binaryDownload.buffer);

  if (actualChecksum !== expectedChecksum.toLowerCase()) {
    throw new Error(`checksum mismatch for ${input.binaryName}: expected ${expectedChecksum}, received ${actualChecksum}`);
  }

  writeFileSync(input.binaryPath, binaryDownload.buffer, { mode: 0o755 });
  chmodSync(input.binaryPath, 0o755);
  writeFileSync(input.versionPath, `${input.version}\n`, 'utf8');
}

function parseChecksum(checksumsText: string, binaryName: string): string {
  for (const line of checksumsText.split('\n')) {
    const trimmed = line.trim();

    if (!trimmed) {
      continue;
    }

    const [checksum, filename] = trimmed.split(/\s+/, 2);

    if (!checksum || !filename) {
      continue;
    }

    const normalizedFilename = basename(filename.replace(/^\*/, ''));

    if (normalizedFilename === binaryName) {
      return checksum.toLowerCase();
    }
  }

  throw new Error(`checksums.txt did not include ${binaryName}`);
}

function sha256(buffer: Buffer): string {
  return createHash('sha256').update(buffer).digest('hex');
}

async function getJson<T>(url: string): Promise<T> {
  const response = await download(url);
  return JSON.parse(response.buffer.toString('utf8')) as T;
}

async function download(url: string, redirects = 0): Promise<DownloadResult> {
  return new Promise((resolve, reject) => {
    const request = get(
      url,
      {
        headers: {
          Accept: 'application/octet-stream, application/json',
          'User-Agent': '@ai-setup/cli bootstrap',
        },
      },
      (response) => {
        const statusCode = response.statusCode ?? 0;
        const location = response.headers.location;

        if (statusCode >= 300 && statusCode < 400 && location) {
          response.resume();

          if (redirects >= MAX_REDIRECTS) {
            reject(new Error(`too many redirects while fetching ${url}`));
            return;
          }

          const nextUrl = new URL(location, url).toString();
          void download(nextUrl, redirects + 1).then(resolve, reject);
          return;
        }

        if (statusCode < 200 || statusCode >= 300) {
          response.resume();
          reject(new Error(`request failed for ${url}: HTTP ${statusCode}`));
          return;
        }

        collectResponse(response).then(
          (buffer) => resolve({ buffer, finalUrl: url }),
          reject,
        );
      },
    );

    request.on('error', reject);
    request.end();
  });
}

async function collectResponse(readable: Readable): Promise<Buffer> {
  const chunks: Buffer[] = [];

  for await (const chunk of readable) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
  }

  return Buffer.concat(chunks);
}

function execCachedBinary(binaryPath: string): Promise<void> {
  return new Promise((resolve) => {
    const child = spawn(binaryPath, process.argv.slice(2), { stdio: 'inherit' });

    child.on('error', (error) => {
      console.error(`ai-setup: failed to execute cached binary: ${formatError(error)}`);
      process.exit(1);
    });

    child.on('close', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
        return;
      }

      process.exit(code ?? 0);
      resolve();
    });
  });
}

function formatError(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return String(error);
}
