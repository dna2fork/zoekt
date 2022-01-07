import os
import re
import time
import sys

from utils import readConfig

def check(scm):
    if scm.exists():
        synced_list = scm.sync()
    else:
        synced_list = scm.clone()
    return synced_list

def main():
    config_filename = os.path.join(os.path.dirname(__file__), 'config')
    config_timestamp = os.path.getmtime(config_filename)
    config = readConfig(config_filename)
    reids_host = config.get('redis') or '127.0.0.1:26379'
    timestamp = 0
    interval = int(config.get('interval', 3600 * 24))

    if len(sys.argv) >= 2:
        projectName = sys.argv[1]
        for scm in config['projects']:
            if scm.getProjectName() != projectName: continue
            check(scm)
        return

    outputProject_filename = config.get('output:project')
    if outputProject_filename:
        f = open(outputProject_filename, 'w+')
        f.close()

    print('[Tips] `touch /tmp/scmsync_exit` for stopping')
    while True:
        if os.path.isfile('/tmp/scmsync_exit'): break
        # update if config change detected
        config_new_timestamp = os.path.getmtime(config_filename)
        if config_new_timestamp != config_timestamp:
            print('[main] config change detected ...')
            config_timestamp = config_new_timestamp
            config = readConfig(config_filename)
            interval = int(config.get('interval', 3600*24))
        # sync code by interval
        if time.time() - timestamp > interval:
            print('[main] start sync projects ...')
            for scm in config['projects']:
                synced_list = check(scm)
                if len(synced_list) > 0:
                    if outputProject_filename:
                        f = open(outputProject_filename, 'a')
                        f.write(scm.getProjectName() + '\n')
                        f.close()
                # deal with `sycned_list`
            timestamp = time.time()
        time.sleep(1)

if __name__ == '__main__':
    main()
