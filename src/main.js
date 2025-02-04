import { intro, outro, cancel, spinner, log, confirm, isCancel } from '@clack/prompts';
import {bgLightCyan, white, green, red, lightCyan, yellow} from 'kolorist';

import {loadConfig, updateConfig} from './config.js';
import { getStagedFiles, stageAllFiles, commitChanges } from './git.js';
import { makeAPIRequest } from './api.js';
import {FLAG_TO_CONFIG_KEY} from "../config/constants.js";

export async function main() {

  const args = process.argv.slice(2);

  for (const arg of args) {
    if (arg.startsWith('--')) {
      const [flagName, value] = arg.slice(2).split('=');
      const configKey = FLAG_TO_CONFIG_KEY[flagName];

      if (configKey) {
        const savedValue = updateConfig(configKey, value || null);

        if (!value) {
          console.log(`${flagName}: ${savedValue || 'not defined'}`);
        }
        process.exit(0);
      }
    }
  }

  const isDryRun = args.includes('--dry');

  const config = loadConfig();

  intro(bgLightCyan(white(' ai-commits ')));
  outro(lightCyan(' Provider: ' + yellow(config['AI_PROVIDER']) || 'not defined'));

  const isStageAll = args.includes('--all');
  if (isStageAll) {
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

    if (isDryRun) {
      s.stop(green('💬 Commit Message:'));
      outro(commitMessage);

      process.exit(0);
    }

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
