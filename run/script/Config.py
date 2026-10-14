"""
    Read ../config.py into memory.
	Use GetConfigDict() to get config.
    main() print the config.
"""
import os
import sys
import socket

def GetParentPath(strPath):
    if not strPath:
        return None;

    lsPath = os.path.split(strPath);
    if lsPath[1]:
        return lsPath[0];

    lsPath = os.path.split(lsPath[0]);
    return lsPath[0];

GAME_HOME = GetParentPath(os.path.dirname(os.path.realpath(__file__)))
def GetConfigDict():
    config_dict={}
    config_file_location = os.path.join(GAME_HOME,"config.py")
    #read config.py file into memory
    if not os.path.exists(config_file_location):
	    sys.stderr.write("config.py not found.\n")
	    exit(1)
    #exec(compile(open(config_file_location,"rb").read(), config_file_location, 'exec'), {socket:socket}, {"config":config_dict})
    exec(compile(open(config_file_location,"rb").read(), config_file_location, 'exec'), {"config":config_dict})

    return config_dict

def main():
    print(GetConfigDict())

if __name__ == "__main__":
    main()
