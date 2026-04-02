import { Command } from 'commander';
import { intro, outro, confirm, log } from '@clack/prompts';
import { Errors, AiSetupError } from '../errors/index.js'
import { fileExists } from '../utils/files.js'
import { readStoreReadonly } from '../store/index.js'
import * as fs from 'fs/promises';
import * as path from 'path';

export function registerEject(program: Command) {
  program
    .command('eject')
    .description('Remove ai-setup management but keep all files')
    .action(async () => {
      await ejectCommand(process.cwd());
    });
}

export async function ejectCommand(targetDir: string) {
  intro('Ejecting from @ricardoborges-teachable/ai-setup');

  const manifestPath = path.join(targetDir, '.ai-setup.json')
  if (!fileExists(manifestPath)) {
    throw Errors.manifestNotFound(targetDir)
  }

  const store = await readStoreReadonly(targetDir)

  const numFiles = store.files.length;

  log.warn(`This will remove the .ai-setup.json manifest.`);
  log.warn(`Your ${numFiles} managed files will be kept, but ai-setup will no longer update them.`);
  
  const shouldEject = await confirm({
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
