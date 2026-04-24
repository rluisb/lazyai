import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import { confirm, intro, log, outro } from '@clack/prompts';
import type { Command } from 'commander';
import { AiSetupError, Errors } from '../errors/index.js'
import { readStoreReadonly } from '../store/index.js'
import { fileExists } from '../utils/files.js'

export function registerEject(program: Command) {
  program
    .command('eject')
    .description('Remove ai-setup management but keep all files')
    .action(async () => {
      await ejectCommand(process.cwd());
    });
}

interface EjectCommandOptions {
  confirmEject?: () => Promise<boolean | symbol>
}

export async function ejectCommand(targetDir: string, opts?: EjectCommandOptions) {
  intro('Ejecting from @ricardoborges-teachable/ai-setup');

  const manifestPath = path.join(targetDir, '.ai-setup.json')
  if (!fileExists(manifestPath)) {
    throw Errors.manifestNotFound(targetDir)
  }

  const store = await readStoreReadonly(targetDir)

  const numFiles = store.files.length;

  log.warn(`This will remove the .ai-setup.json manifest.`);
  log.warn(`Your ${numFiles} managed files will be kept, but ai-setup will no longer update them.`);
  
  const shouldEject = opts?.confirmEject
    ? await opts.confirmEject()
    : await confirm({
        message: 'Are you sure you want to eject?',
        initialValue: false
      });

  if (!shouldEject || typeof shouldEject === 'symbol') {
    outro('Eject cancelled');
    return;
  }

  try {
    await fs.unlink(manifestPath);
    log.success('Successfully removed .ai-setup.json');
    outro('Eject complete. You now fully own all files.');
  } catch (err: unknown) {
    if (err instanceof AiSetupError) throw err
    throw Errors.unknown(err instanceof Error ? err.message : String(err))
  }
}
