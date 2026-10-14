"""
    This python script auto scan all the ".tmpl" files in "." directory.
    Then use "config.py" to replace all the variable and remove the ".tmpl" suffix.
"""
import os,sys
import re
import getopt
import shutil
import glob
import Config
from MakeTmpl import TmplFile

config_dict = Config.GetConfigDict()
VariableRegex = re.compile("\$\{\w*\}")
BusIDRegex = re.compile("\d+.\d+.\d+.\d+")

def CheckAllTmplFile():
    #check all variable is in config.py
    for root, dirs, files in os.walk(Config.GAME_HOME):
        for file in files:
            if os.path.splitext(file)[1] == ".tmpl":
                fileHandle = open(os.path.join(root,file),"r")
                content = fileHandle.read()
                fileHandle.close()
                for eachVariable in VariableRegex.findall(content):
                    #if not config_dict.has_key(eachVariable.strip("${}")):
                    if eachVariable.strip("${}") not in config_dict:
                        sys.stderr.write("in %s/%s:" % (root,file))
                        sys.stderr.write("\"%s\" not defined in config.py.\n" % eachVariable.strip("${}"))
                        exit(1)

def MakeTmpl(is_verbose):
    #replace all the variable and remove the ".tmpl" suffix.
    for root, dirs, files in os.walk(Config.GAME_HOME):
        target_dirs = []
        target_bus_id = []
        for file in files:
            if os.path.splitext(file)[1] == ".tmpl":
                tmpl_file = TmplFile(config_dict,os.path.join(root,file),is_verbose,target_dirs,target_bus_id)
                tmpl_file.Flush()



def main():
    #print("Start generate config")
    is_verbose = False
    optlist, args = getopt.getopt(sys.argv[1:],'', ['verbose','clear','force'])
    for (opt, val) in optlist:
        if opt == '--verbose':
            is_verbose = True

    replace_config_dict = {}
    for arg in args:
        if '=' in arg:
            replace_config_dict[arg.split("=")[0]] = arg.split("=")[1]

    for each_config_key,each_config_value in replace_config_dict.items():
        config_dict[each_config_key] = each_config_value

    # do tmpl parsing
    CheckAllTmplFile()
    #print("CheckTmplFile tcm config")
    MakeTmpl(is_verbose)
   
if __name__ == "__main__":
    main()
