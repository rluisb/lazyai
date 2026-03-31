import { Command } from 'commander';
import { intro, outro, confirm, log } from '@clack/prompts';
import { readManifest } from '../utils/manifest.js';
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
  intro('Ejecting from @teachable/ai-setup');
  
  const manifest = await readManifest(targetDir);
  
  if (!manifest) {
    log.info('No .ai-setup.json manifest found. Project is not managed by ai-setup.');
    outro('Eject complete');
    return;
  }

  const numFiles = Object.keys(manifest.files || {}).length;
  
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
    const manifestPath = path.join(targetDir, '.ai-setup.json');
    await fs.unlink(manifestPath);
    log.success('Successfully removed .ai-setup.json');
    outro('Eject complete. You now fully own all files.');
  } catch (err: any) {
    log.error(`Failed to eject: ${err.message}`);
    outro('Eject failed');
  }
}
