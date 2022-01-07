import os
import subprocess

class Git(object):
    def __init__(self, url, path, project, branch, username, flags=None):
        self._url = url
        self._username = username
        self._path = path
        self._project = project
        self._branch = branch
        self._flags = flags
        self._parsedFlagOptions = self._parseFlags(flags)

    def _parseFlags(self, flags):
        options = {}
        if not flags: return options
        if 'h': options['history'] = True
        return options

    def _removeEmptyTail(self, array):
        if len(array) == 0: return
        if array[-1]: return
        array.pop()

    def _diffPath(self, path):
        parts = path.split(' ')[2:]
        n = int(len(parts) / 2)
        if n == 0: return None
        path = ' '.join(parts[n:])[1:] # b/path/to/file -> /path/to/file
        return path

    def getProjectName(self):
        return self._project

    def getBranchName(self):
        return self._branch

    def getSourceRoot(self):
        return self._path

    def getProjectPathObject(self, filename):
        if not filename: return None
        prefix = self.getSourceRoot()
        if not filename.startswith(prefix): return None
        filepath = filename[len(prefix):]
        return { "project": self.getProjectName(), "path": filepath }

    def exists(self):
        return os.path.isdir(self._path)

    def actFlags(self, synced_list):
        if not self._parsedFlagOptions: return
        print('[git] additional analysis ...')
        if self._parsedFlagOptions['history']:
            self.doHistoryAnalysis(synced_list)

    def doHistoryAnalysis(self, synced_list):
        try:
            cmd = 'cd {0}; mkdir -p ./.git/.zoekt/commit'.format(self._path)
            subprocess.check_output(cmd, shell=True)
            cmd = 'cd {0}; git log --pretty=format:%H'.format(self._path)
            commit_list = subprocess.check_output(cmd, shell=True).decode().split('\n')
            print('[git] commits={0}'.format(len(commit_list)))
            for commit in commit_list:
                if not commit: continue
                try:
                    if os.path.isfile('{0}/.git/.zoekt/commit/{1}'.format(self._path, commit)): continue
                    cmd = 'cd {0}; git log --name-status --diff-filter="ACDMRT" -1 -U {1} > ./.git/.zoekt/commit/{1}'.format(self._path, commit)
                    subprocess.check_output(cmd, shell=True).decode()
                except:
                    print('[git] failed to fetch commit {0}'.format(commit))
            print('[git] history analysis complete')
        except:
            print('[git] history analysis failed')

    def clone(self):
        print('[git] clone {0} branch={1}'.format(self._path, self._branch))
        try:
            cmd = 'git clone \'{0}\' -b {2} {1}'.format(self._url, self._path, self._branch)
            subprocess.check_output(cmd, shell=True)
            cmd = 'find {0} -type f | grep -v -e "[.]git/"'.format(self._path)
            synced_list = subprocess.check_output(cmd, shell=True).decode().split('\n')
            self._removeEmptyTail(synced_list)
        except:
            synced_list = []
        self.actFlags(synced_list)
        print('[git] clone complete')
        return synced_list

    def sync(self):
        print('[git] sync {0}'.format(self._path))
        try:
            cmd = 'cd {0}; git branch | grep "* "'.format(self._path)
            branch = subprocess.check_output(cmd, shell=True).decode()
            branch = branch.split(' ')[-1].strip()
            cmd = 'cd {0}; git fetch --all; git diff HEAD "origin/{1}" | grep -e "^diff --git "'.format(self._path, branch)
            synced_list = subprocess.check_output(cmd, shell=True).decode().split('\n')
            self._removeEmptyTail(synced_list)
            synced_list = list(filter(None, map(self._diffPath, synced_list)))
            cmd = 'cd {0}; git reset --hard "origin/{1}"'.format(self._path, branch)
            subprocess.check_output(cmd, shell=True)
        except:
            synced_list = []
        self.actFlags(synced_list)
        print('[git] sync complete')
        return synced_list

    def __str__(self):
        return '[scm/git] {url={0},path={1},branch={2},user={3}}'.format(self._url, self._path, self._branch, self._username)

    def __repr(self):
        return self.__str__()
