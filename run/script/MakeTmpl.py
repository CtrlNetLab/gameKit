"""
    .tmpl file parse class.
    main() is the testcase.
"""
import os,sys
import subprocess
import re,copy
import Config
import MakeWorld
from io import StringIO

VariableRegex = re.compile("\$\{\w*\}")
ForeachStartRegex = re.compile("\[\[foreach")
PythonScriptStartRegex = re.compile("\[\[python")
BracketsEndRegex = re.compile("\B\]\]")

class TmplFile:
    """
        parse .tmpl file
    """
    def __init__(self,config_dict,file,is_verbose,target_dirs,target_bus_id):
        self.filepath = os.path.dirname(file)
        self.filename = os.path.basename(file)
        fileHandle = open(file,"r")
        self.lines = fileHandle.readlines()
        fileHandle.close()
        self.is_verbose = is_verbose
        self.target_dirs = target_dirs  #TODO this may be merged
        self.target_bus_id = target_bus_id  #TODO this may be merged
        self.config_dict = config_dict
        self.server_id = int(config_dict[MakeWorld.serverid_str])
        #self.i_battle_world_id = int(config_dict[MakeWorld.battle_world_str]) 

    def __Print(self,str):
        if self.is_verbose:
            print(str)

    def __MakeLine(self,linenum,line,bus_id):
        if bus_id != "":
            world_id = bus_id.split(".")[0]
            zone_id = bus_id.split(".")[1]
            function_id = bus_id.split(".")[2]
            instance_id = bus_id.split(".")[3]
            if WorldRegex.findall(line):
                line = line.replace("[[world]]",world_id)
            if ZoneRegex.findall(line):
                line = line.replace("[[zone]]",zone_id)
            if FunctionRegex.findall(line):
                line = line.replace("[[function]]",function_id)
            if InstanceRegex.findall(line):
                line = line.replace("[[instance]]",instance_id)
        #if self.i_world_id > 255 or self.i_battle_world_id > 255:
        #    line = MakeWorld.replace_worldid(line,self.config_dict)
        for eachVariable in VariableRegex.findall(line):
            replaceWord = eachVariable.strip("${}")
            beReplaceWord=str(self.config_dict[replaceWord])
            self.__Print("line num:%s    %s ====> %s" % (str(linenum+1),replaceWord,beReplaceWord))
            line = line.replace(eachVariable,beReplaceWord)
        return line

    def __ExtendForeachKeyWord(self,lines,parseStartLineNo,parseEndLineNo):
        #print self.filename,parseStartLineNo,parseEndLineNo
        foreachKeyword = self.__MakeLine(parseStartLineNo,lines[parseStartLineNo].split("[[")[1].strip("\r\n"),"")
        foreachSourceLines = lines[parseStartLineNo+1:parseEndLineNo]
        repeatCount = int(foreachKeyword.split("in")[1].strip(" "))
        repeatArg = foreachKeyword.split("in")[0].split("foreach")[1].strip(" ")
        del lines[parseStartLineNo:parseEndLineNo+1]
        for i in xrange(repeatCount):
            beReplaceWord = str(i+1)
            for line in foreachSourceLines:
                line = line.replace("[["+repeatArg+"]]",beReplaceWord)
                lines.insert(parseStartLineNo,line)
                parseStartLineNo += 1
    
    def __ExtendPythonKeyWord(self,lines,parseStartLineNo,parseEndLineNo,bus_id):
        #print bus_id
        bus_id_list = bus_id.split(".")
        myConfig = {}
        if len(bus_id_list) == 4:
            myConfig = {"world":bus_id_list[0],"zone":bus_id_list[1],"function":bus_id_list[2],"instance":bus_id_list[3]}
        pythonSourceLines = lines[parseStartLineNo+1:parseEndLineNo]
        pythonSource = "\n".join([line.strip("\n\r") for line in pythonSourceLines])
        del lines[parseStartLineNo:parseEndLineNo+1]
        buff = StringIO()
        temp = sys.stdout
        sys.stdout = buff
        #print myConfig
        exec(pythonSource, {"config":self.config_dict,"this_file":myConfig})
        sys.stdout = temp
        lines.insert(parseStartLineNo,buff.getvalue())

    def __Parse(self,lines,bus_id=""):
        self.__Print("parsing %s/%s" % (self.filepath,self.filename))
        # scan for [[]]
        parsingForeach = False
        parsingPythonScript = False
        parseStartLineNo = 0
        parseEndLineNo = 0
        parseLinePairList = []
        for linenum, line in enumerate(lines):
            if ForeachStartRegex.findall(line):
                if parsingForeach == False and parsingPythonScript == False:
                    parsingForeach = True
                    parseStartLineNo = linenum
                else:
                    sys.stderr.write("in %s/%s:" % (self.filepath,self.filename))
                    sys.stderr.write("line num:%s syntex error!\n" % str(linenum+1))
                    exit(1)
            if PythonScriptStartRegex.findall(line):
                if parsingForeach == False and parsingPythonScript == False:
                    parsingPythonScript = True
                    parseStartLineNo = linenum
                else:
                    sys.stderr.write("in %s/%s:" % (self.filepath,self.filename))
                    sys.stderr.write("line num:%s syntex error!\n" % str(linenum+1))
                    exit(1)
            if BracketsEndRegex.findall(line):
                if parsingForeach == True:
                    parseEndLineNo = linenum
                    parseLinePairList.insert(0,(parseStartLineNo,parseEndLineNo,"foreach"))
                    #can't extend here,because line index may change
                    #self.__ExtendForeachKeyWord(lines,parseStartLineNo,parseEndLineNo)
                    parsingForeach = False
                elif parsingPythonScript == True:
                    parseEndLineNo = linenum
                    parseLinePairList.insert(0,(parseStartLineNo,parseEndLineNo,"python"))
                    #can't extend here,because line index may change
                    #self.__ExtendPythonKeyWord(lines,parseStartLineNo,parseEndLineNo)
                    parsingPythonScript = False
                else:
                    sys.stderr.write("in %s/%s:" % (self.filepath,self.filename))
                    sys.stderr.write("line num:%s syntex error!\n" % str(linenum+1))
                    exit(1)
        
        for (startLineNo,EndLineNo,KeywordName) in parseLinePairList:
            if KeywordName == "foreach":
                self.__ExtendForeachKeyWord(lines,startLineNo,EndLineNo)
            elif KeywordName == "python":
                self.__ExtendPythonKeyWord(lines,startLineNo,EndLineNo,bus_id)
        
        for linenum, line in enumerate(lines):
            lines[linenum] = self.__MakeLine(linenum,lines[linenum],bus_id)

    def __FlushTo(self,resultFileName):
        #if contain ".sh" do else
        if os.path.splitext(resultFileName)[1] != ".sh" and len(self.target_bus_id) != 0:
            for bus_id in self.target_bus_id:
                #if busid in file name is not current,not copy
                if BusIDRegex.findall(resultFileName):
                    if BusIDRegex.findall(resultFileName)[0] != bus_id:
                        continue

                lines = copy.copy(self.lines)
                self.__Parse(lines,bus_id)
                #test
                if os.path.splitext(resultFileName)[1] != ".xml":
                    target_file_name = os.path.splitext(resultFileName)[0] + ".xml"
                else:
                    target_file_name = os.path.splitext(resultFileName)[0] + "_" + bus_id + os.path.splitext(resultFileName)[1]

                target_file_path = os.path.dirname(target_file_name)
                if not os.path.exists(target_file_path):
                    os.makedirs(target_file_path)
                if os.path.exists(target_file_name):
                    os.chmod(target_file_name,0o0644)
                newFile = open(target_file_name,"w")
                newFile.write("".join([line.decode("GBK").encode("GBK") for line in lines]))
                self.__Print("write to %s" % target_file_name)
                newFile.close()

                # the generated file do not allowed to be change 
                #.sh file set to executable
                if os.path.splitext(target_file_name)[1] == ".sh":
                    os.chmod(target_file_name,0o0555)
                #.xml file set to readonly
                if os.path.splitext(target_file_name)[1] == ".xml":
                    os.chmod(target_file_name,0o0444)
        else:
            lines = copy.copy(self.lines)
            self.__Parse(lines)
            target_file_name = resultFileName
            target_file_path = os.path.dirname(target_file_name)
            if not os.path.exists(target_file_path):
                os.makedirs(target_file_path)
            if os.path.exists(target_file_name):
                os.chmod(target_file_name,0o0644)
            newFile = open(target_file_name,"w")
            # newFile.write("".join([line.decode("GBK").encode("GBK") for line in lines]))
            newFile.write("".join([line for line in lines]))
            self.__Print("write to %s" % target_file_name)
            newFile.close()

            # the generated file do not allowed to be change 
            #.sh file set to executable
            if os.path.splitext(target_file_name)[1] == ".sh":
                os.chmod(target_file_name,0o0555)
                #os.system('dos2unix %s' % target_file_name)
                #commands.getoutput('dos2unix %s' % target_file_name)
            #.xml file set to readonly
            if os.path.splitext(target_file_name)[1] == ".xml":
                os.chmod(target_file_name,0o0444)

        self.__Print("make %s done" % resultFileName)
        self.__Print("")

    def Flush(self):
        # check if xxx.developer.dev.tmpl or xxx.developer.idc.tmpl or xxx.business.idc.tmpl
        filename_sequence = self.filename.split(".")
        if "developer" in filename_sequence or "business" in filename_sequence or "formaltest" in filename_sequence:
            if self.config_dict["ServerMode"] in filename_sequence:
                if self.config_dict["OpenPF"] in filename_sequence:
                        filename_sequence.remove(self.config_dict["ServerMode"])
                        filename_sequence.remove(self.config_dict["OpenPF"])
                        #print(self.filename)
                else:               
                                        return
            else:
                return
        if "ios" in filename_sequence or "android" in filename_sequence:
            if self.config_dict["ClientPlatform"] in filename_sequence:
                filename_sequence.remove(self.config_dict["ClientPlatform"])
            else:
                return
                
        
        self.filename = ".".join(filename_sequence)
        #print(self.filename)             
        if len(self.target_dirs) == 0:
            self.__FlushTo(os.path.join(self.filepath,os.path.splitext(self.filename)[0]))
        else:
            self.__FlushTo(os.path.join(self.filepath,os.path.splitext(self.filename)[0]))
            for target_path in self.target_dirs:
                self.__FlushTo(os.path.join(target_path,os.path.splitext(self.filename)[0]))


def main():
    config_dict = Config.GetConfigDict()
    is_verbose = True

if __name__ == "__main__":
    main()

