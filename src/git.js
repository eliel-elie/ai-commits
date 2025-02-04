import { execSync } from 'child_process';
import { red } from 'kolorist';

export function stageAllFiles() {
  const isWindows = process.platform === 'win32';
  const command = isWindows ? 'git add . 2>NUL' : 'git add . 2>/dev/null';
  execSync(command, { stdio: 'inherit' });
}

export function getStagedFiles(type = 'names') {
  try {
    if (type === 'names') {
      const files = execSync('git diff --staged --name-only').toString().trim().split('\n');
      return files.filter(file => file.length > 0);
    }
    
    if (type === 'diff') {
      return execSync('git diff --staged').toString();
    }
  } catch (error) {
    console.error(red('Error retrieving staged files:', error.message));
    process.exit(1);
  }
}

export function commitChanges(message) {
  try {
    execSync(`git commit --no-verify -m "${message.replace(/"/g, '\\"')}"`, { stdio: 'inherit' });
  } catch (error) {
    console.error(red('Error committing changes:', error.message));
    process.exit(1);
  }
}
