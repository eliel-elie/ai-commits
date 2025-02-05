import { intro, outro, cancel, spinner, log, confirm, isCancel } from '@clack/prompts';
import {bgLightCyan, white, green, red, lightCyan, yellow} from 'kolorist';

import {loadConfig, updateConfig} from './config.js';
import { getStagedFiles, stageAllFiles, commitChanges } from './git.js';
import { makeAPIRequest } from './api.js';
import {FLAG_TO_CONFIG_KEY} from "../config/constants.js";
import {t} from "./i18n/index.js";

export async function main() {

  const args = process.argv.slice(2);

  for (const arg of args) {
    if (arg.startsWith('--')) {
      const [flagName, value] = arg.slice(2).split('=');
      const configKey = FLAG_TO_CONFIG_KEY[flagName];

      if (configKey) {
        const savedValue = updateConfig(configKey, value || null);

        if (!value) {
          console.log(`${flagName}: ${savedValue || t('cli.notConfigured')}`);
        }
        process.exit(0);
      }
    }
  }

  const isDryRun = args.includes('--dry');

  const config = loadConfig();

  intro(bgLightCyan(white(t('cli.intro'))));
  outro(lightCyan(' Provider: ' + yellow(config['AI_PROVIDER']) || 'not defined'));

  const isStageAll = args.includes('--all');
  if (isStageAll) {
    outro(t('cli.stagingAll'));
    stageAllFiles();
  }

  const detectingFiles = spinner();
  detectingFiles.start(t('cli.detectingFiles'));

  const diff = await getStagedFiles('diff');
  if (!diff) {
    log.error(red(t('cli.noChanges')));
    process.exit(0);
  }

  const stagedFiles = await getStagedFiles('names');
  detectingFiles.stop(t('cli.filesDetected', { count : stagedFiles.length}));

  stagedFiles.forEach(file => {
    console.log(`   ${file}`);
  });

  const s = spinner();
  s.start(t('cli.analyzing'));
  const commitMessage = await makeAPIRequest(diff);
  s.stop(t('cli.changesAnalyzed'));

  if (commitMessage) {

    if (isDryRun) {
      s.stop(green(t('dryRun.commitMessage')));
      outro(commitMessage);

      process.exit(0);
    }

    const confirmed = await confirm({
      message: t('cli.useMessage') + `\n\n      ${commitMessage}\n`,
      active: t('confirm.yes'),
      inactive: t('confirm.no'),
    });

    if (!confirmed || isCancel(confirmed)) {
      cancel(t('cli.commitCancelled'));
      return;
    }

    commitChanges(commitMessage);
    outro(`${green('✔')} ` + t('cli.commitSuccess'));
  } else {
    log.error(red(t('cli.errorGenerating')));
    process.exit(1);
  }
}
