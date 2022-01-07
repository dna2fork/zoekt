import os
import subprocess

class P4(object):
    def __init__(self, host, client, project, username, flags=None):
        self._url = host
        self._username = username
        self._path = client
        self._project = project
        self._client = self._p4client()
        self._sourceroot = os.path.abspath(os.path.join(self._client.get('root'), self._project))
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

    def _extractPath(self, path):
        # e.g. //depot/a#2 - deleted as /path/to/a
        #      //depot/b#1 - added as /path/to/b
        parts = path.split(' ')
        return parts[-1]

    def getProjectName(self):
        return self._project

    def getSourceRoot(self):
        return self._sourceroot

    def getProjectPathObject(self, filename):
        if not filename: return None
        prefix = self.getSourceRoot()
        if not filename.startswith(prefix): return None
        filepath = filename[len(prefix):]
        return { "project": self.getProjectName(), "path": filepath }

    def exists(self):
        path = self.getSourceRoot()
        return os.path.isdir(path)

    def actFlags(self, synced_list):
        if not self._parsedFlagOptions: return
        print('[p4] additional analysis ...')
        if self._parsedFlagOptions['history']:
            self.doHistoryAnalysis(synced_list)

    def _p4remote2local(self, client, filepath):
        for view in client['view']:
            remote = view[0]
            local = view[1]
            parts = filepath.split(remote)
            if len(parts) > 1 and parts[0] == '':
                return local + remote.join(parts[1:])
        return None

    def _p4describe(self, client, changeObj):
        projectPath = self.getSourceRoot()
        if not projectPath.endswith('/'): projectPath += '/'
        try:
            cmd = 'P4USER={2} P4PORT={0} P4CLIENT={1} p4 describe -s {3}'.format(self._url, self._path, self._username, changeObj['changeId'])
            description = subprocess.check_output(cmd, shell=True).decode()
            parts = description.split('Affected files ...')
            commitMessage = 'Affected files ...'.join(parts[:-1])
            affectedFileList = parts[-1].strip().split('\n')
            filtered = []
            for fileDesc in affectedFileList:
                if not fileDesc: continue
                try:
                    i = fileDesc.index(' ')
                    j = fileDesc.rindex(' ')
                    if i < j:
                        filepath = fileDesc[i+1:j]
                        local_filename = self._p4remote2local(client, filepath)
                        # maybe the modified file is not in this project
                        if not local_filename: continue
                        parts = local_filename.split(projectPath)
                        if len(parts) > 1 and parts[0] == '':
                            filtered.append(fileDesc)
                except:
                    continue
            filteredDescription = commitMessage + '\nAffected files ...\n\n' + '\n'.join(filtered)
            return filteredDescription
        except:
            return None

    def _p4local2remote(self, client, filename):
        for view in client['view']:
            remote = view[0]
            local = view[1]
            parts = filename.split(local)
            if len(parts) > 1 and parts[0] == '':
                return remote + local.join(parts[1:])
        return None

    def _p4filelog(self, client, filename):
        filename = os.path.abspath(filename)
        if os.path.isfile(filename):
            # file added/modified
            cmd = 'P4USER={2} P4PORT={0} P4CLIENT={1} p4 filelog -s {3}'.format(self._url, self._path, self._username, filename)
        elif client['viewMap']:
            # file deleted
            remote_filename = self._p4local2remote(client, filename)
            cmd = 'P4USER={2} P4PORT={0} P4CLIENT={1} p4 filelog -s {3}'.format(self._url, self._path, self._username, remote_filename)
        try:
            changeList = subprocess.check_output(cmd, shell=True).decode().split('\n')
            changeObjList = []
            for change in changeList:
                if not change: continue
                parts = change.split(' ')
                # skip the first line that show remote file path
                if (len(parts) < 5): continue
                revision = parts[1]
                if revision == '...':
                    lastObj = changeObjList[-1]
                    if lastObj:
                        lastObj['additionalMessage'] = change
                else:
                    change_id = parts[3]
                    change_type = parts[4]
                    changeObjList.append({
                        'revision': revision,
                        'changeId': change_id,
                        'type': change_type
                    })
            return changeObjList
        except:
            return []

    def _p4client(self):
        client = {}
        try:
            cmd = 'P4USER={2} P4PORT={0} P4CLIENT={1} p4 client -o'.format(self._url, self._path, self._username)
            clientLines = subprocess.check_output(cmd, shell=True).decode().split('\n')
            key = None
            value = None
            for line in clientLines:
                lineOrigin = line
                line = line.strip()
                if not line: continue
                if line.startswith('#'): continue
                if line.startswith('Client:'):
                    key = 'client'
                    if len(line) > len('Client:'):
                        value = line[len('Client:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('Update:'):
                    key = 'update'
                    if len(line) > len('Update:'):
                        value = line[len('Update:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('Access:'):
                    key = 'access'
                    if len(line) > len('Access:'):
                        value = line[len('Access:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('Owner:'):
                    key = 'owner'
                    if len(line) > len('Owner:'):
                        value = line[len('Owner:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('Description:'):
                    key = 'description'
                    client[key] = ''
                    if len(line) > len('Description:'):
                        value = line[len('Description:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('Root:'):
                    key = 'root'
                    if len(line) > len('Root:'):
                        value = line[len('Root:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('Options:'):
                    key = 'options'
                    if len(line) > len('Options:'):
                        value = line[len('Options:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('SubmitOptions:'):
                    key = 'submitoptions'
                    if len(line) > len('SubmitOptions:'):
                        value = line[len('SubmitOptions:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('LineEnd:'):
                    key = 'lineend'
                    if len(line) > len('LineEnd:'):
                        value = line[len('LineEnd:'):].strip()
                        client[key] = value
                        key = None
                elif line.startswith('View:'):
                    key = 'view'
                    client[key] = []
                elif key == 'description':
                    client[key] += lineOrigin + '\n'
                elif key == 'view':
                    if not line.startswith('//'): continue
                    parts = line.split('...')
                    remote = parts[0].strip()
                    local = parts[1].strip()
                    client[key].append([remote, local])
            if 'view' in client and 'root' in client:
                client['viewMap'] = True
                for view in client['view']:
                    remote = view[0]
                    local = view[1]
                    localParts = [client['root']] + local.split('/')[3:]
                    local = '/'.join(localParts)
                    view[1] = local
            else:
                client['viewMap'] = False
                print('[p4] p4client: cannot track deleted files ...')
        except:
            pass
        return client

    def doHistoryAnalysis(self, synced_list):
        try:
            client = self._client
            cmd = 'cd {0}; mkdir -p ./.p4/.zoekt/commit'.format(self.getSourceRoot())
            subprocess.check_output(cmd, shell=True)
            processedCount = 0
            for filename in synced_list:
                changeObjList = self._p4filelog(client, filename)
                for changeObj in changeObjList:
                    changeId = changeObj['changeId']
                    filename = os.path.join(self.getSourceRoot(), '.p4', '.zoekt', 'commit', '{0}'.format(changeId))
                    try:
                        if os.path.isfile(filename):
                            cmd = 'cat {0}'.format(filename)
                            text = subprocess.check_output(cmd, shell=True).decode()
                            if 'additionalMessage' in changeObj and changeObj['additionalMessage'] not in text:
                                f = open(filename, 'a')
                                f.write('\n' + changeObj['additionalMessage'])
                                f.close()
                        else:
                            description = self._p4describe(client, changeObj)
                            f = open(filename, 'w+')
                            f.write(description)
                            if 'additionalMessage' in changeObj:
                                f.write('\n' + changeObj['additionalMessage'])
                            f.close()
                    except:
                        print('[p4] cannot process Change {0}'.format(changeId))
                processedCount += 1
                if processedCount % 1000 == 0:
                    N = len(synced_list)
                    print('[p4] history analysis: {0}/{1}, {2}%'.format(processedCount, N, processedCount*1000/N/100.0))
            print('[p4] history analysis complete')
        except:
            print('[p4] history analysis failed')

    def clone(self):
        print('[p4] clone {0}/{1}'.format(self._url, self._path))
        try:
            cmd = 'P4USER={2} P4PORT={0} P4CLIENT={1} p4 sync -f'.format(self._url, self._path, self._username)
            subprocess.check_output(cmd, shell=True)
            path = self.getSourceRoot()
            if not path: return []
            cmd = 'find {0} -type f'.format(path)
            synced_list = subprocess.check_output(cmd, shell=True).decode().split('\n')
            self._removeEmptyTail(synced_list)
        except:
            synced_list = []
        self.actFlags(synced_list)
        print('[p4] clone complete')
        return synced_list

    def sync(self):
        print('[p4] sync {0}/{1}'.format(self._url, self._path))
        try:
            cmd = 'P4USER={2} P4PORT={0} P4CLIENT={1} p4 sync'.format(self._url, self._path, self._username)
            synced_list = subprocess.check_output(cmd, shell=True).decode().split('\n')
            self._removeEmptyTail(synced_list)
            synced_list = list(map(self._extractPath, synced_list))
        except:
            synced_list = []
        self.actFlags(synced_list)
        print('[p4] sync complete')
        return synced_list

    def __str__(self):
        return '[scm/p4] {host={0},client={1},user={2}}'.format(self._url, self._path, self._username)

    def __repr(self):
        return self.__str__()
