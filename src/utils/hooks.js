import { execSync } from 'child_process';
import { red, yellow } from 'kolorist';
import {log} from "@clack/prompts";
import {t} from "../i18n/index.js";

export function hasPreCommitHook() {
    try {
        const hookPath = '.git/hooks/pre-commit';
        const isWindows = process.platform === 'win32';
        const command = isWindows ? `if exist "${hookPath}" echo true` : `test -f "${hookPath}" && echo true`;

        const result = execSync(command).toString().trim();
        return result === 'true';
    } catch (error) {
        return false;
    }
}

export function runPreCommitHook() {
    try {
        const stagedFiles = execSync('git diff --staged --name-only').toString().trim();

        execSync('.git/hooks/pre-commit', { stdio: 'inherit' });

        const currentStaged = execSync('git diff --staged --name-only').toString().trim();

        if (currentStaged !== stagedFiles) {
            log.warn(yellow(t('git.reStaging')));
            stagedFiles.split('\n').forEach(file => {
                if (file) {
                    execSync(`git add "${file}"`, { stdio: 'inherit' });
                }
            });
        }

        return true;
    } catch (error) {
        log.error(red(t('git.error'), error.message));
        return false;
    }
}