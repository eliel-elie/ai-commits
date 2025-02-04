import { intro, outro, cancel, spinner, log, confirm, isCancel } from '@clack/prompts';
import {bgLightCyan, white, green, red, lightCyan, yellow} from 'kolorist';

import {handleConfig, loadConfig} from './config.js';
import { getStagedFiles, stageAllFiles, commitChanges } from './git.js';
import { makeAPIRequest } from './api.js';

export async function main() {
  if (process.argv[2] === 'config') {
    return handleConfig(process.argv);
  }

  const config = loadConfig();

  intro(bgLightCyan(white(' ai-commits ')));
  outro(lightCyan(' Provider: ' + yellow(config['AI_PROVIDER'])));

  const allFlag = process.argv.includes('-a') || process.argv.includes('--all');
  if (allFlag) {
    outro('Staging all changes');
    stageAllFiles();
  }

  const detectingFiles = spinner();
  detectingFiles.start('Detecting staged files');

  const diff = await getStagedFiles('diff');
  if (!diff) {
    log.error(red("No staged changes found. Stage your changes manually, or use the `--all` flag."));
    process.exit(0);
  }

  const stagedFiles = await getStagedFiles('names');
  detectingFiles.stop(`Detected ${stagedFiles.length} staged file${stagedFiles.length > 1 ? 's' : ''}:`);

  stagedFiles.forEach(file => {
    console.log(`   ${file}`);
  });

  const s = spinner();
  s.start('The AI is analyzing your changes');
  const commitMessage = await makeAPIRequest(diff);
  s.stop('Changes analyzed');

  if (commitMessage) {
    const confirmed = await confirm({
      message: `Use this commit message?\n\n      ${commitMessage}\n`,
    });

    if (!confirmed || isCancel(confirmed)) {
      cancel('Commit cancelled');
      return;
    }

    commitChanges(commitMessage);
    outro(`${green('✔')} Successfully committed!`);
  } else {
    log.error(red("Error: Could not generate commit message."));
    process.exit(1);
  }
}
