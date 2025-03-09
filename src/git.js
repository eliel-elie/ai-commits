import { execSync } from 'child_process';
import { red } from 'kolorist';
import {t} from "./i18n/index.js";
import {log} from "@clack/prompts";
import {hasPreCommitHook, runPreCommitHook} from "./utils/hooks.js";

export function stageAllFiles() {
  const isWindows = process.platform === 'win32';
  const command = isWindows ? 'git add . 2>NUL' : 'git add . 2>/dev/null';
  execSync(command, { stdio: 'inherit' });
}

export function stageFile(path) {
  const isWindows = process.platform === 'win32';
  const command = isWindows ? `git add "${path}" 2>NUL` : `git add "${path}" 2>/dev/null`;
  execSync(command, { stdio: 'inherit' });
}

export function getStagedFiles(type = 'names', specificFile = null) {
  try {

    if (specificFile) {

      if (type === 'names') {
        return [specificFile];
      }

      if (type === 'diff') {
        return execSync(`git diff --staged -- "${specificFile}"`).toString();
      }

    } else {

      if (type === 'names') {
        const files = execSync('git diff --staged --name-only').toString().trim().split('\n');
        return files.filter(file => file.length > 0);
      }

      if (type === 'diff') {
        return execSync('git diff --staged').toString();
      }

    }

  } catch (error) {
    log.error(red(t('git.errorStagedFiles'), error.message));
    process.exit(1);
  }
}

export function commitChanges(message, skipVerify = false, specificFile = null) {
  try {

    if (!skipVerify && hasPreCommitHook()) {
      const hookSuccess = runPreCommitHook();
      if (!hookSuccess) {
        log.error(red(t('git.failedPreCommit')));
        process.exit(1);
      }
    }

    let command;
    if (specificFile) {

      command = skipVerify
          ? `git commit --no-verify -m "${message.replace(/"/g, '\\"')}" "${specificFile}"`
          : `git commit -m "${message.replace(/"/g, '\\"')}" "${specificFile}"`;

    } else {

      command = skipVerify
          ? `git commit --no-verify -m "${message.replace(/"/g, '\\"')}"`
          : `git commit -m "${message.replace(/"/g, '\\"')}"`;

    }

    execSync(command, { stdio: 'inherit' });

  } catch (error) {
    log.error(red(t('git.errorCommit'), error.message));
    process.exit(1);
  }
}
