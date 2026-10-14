


$CurrentDirectory = "$PSScriptRoot/../protocol/configs"

$PROTO_PREFIX = "./"
$DATA_PREFIX = "../../binconf/configs"
$OUTDIR = "../../pkg/configs"

$PrimaryKeyFile = "$CurrentDirectory/primarykey.txt"

$PrimaryKeyContent = Get-Content $PrimaryKeyFile

$PrimaryKeyContent | ForEach-Object {
    # 判断元素中是否包含 "Config" 字符串
    $element = $_
    if ($element -match "Config") {
        $line = $element -split ":"
        $Dataconfig = $line[0].Trim()
        $Key = $line[1].Trim()
        $Type = $line[2].Trim()

        $ProtoFile = Join-Path $PROTO_PREFIX "$Dataconfig.proto"
        $ProtoFile = "$CurrentDirectory\$ProtoFile"

        if (Test-Path $ProtoFile -PathType Leaf) {
            $Struct = Get-Content ${ProtoFile} | Select-String -Pattern '^message ([^{ \t]*)' | Select-Object -First 1 | ForEach-Object {
                $_.Matches.Groups[1].Value
            }
            $Res = ($Struct.Substring(0, 1).ToUpper()) + $Struct.Substring(1)
            $ConfigFile = Join-Path $DATA_PREFIX "$Dataconfig.json"
            $ConfigFile = $ConfigFile -replace '\\', '/'
            $PrimaryKey=$Key
            $OUTFILE = Join-Path $OUTDIR ("${Res}ResMgr.pb.go")
            $OUTFILE = "$CurrentDirectory\$OUTFILE"

            Copy-Item "$CurrentDirectory\ResMgr.go.tmpl" $OUTFILE -Force

#            Write-Output ===============>
#            Write-Output $OUTFILE
#            Write-Output $Res
#            Write-Output $ConfigFile
#            Write-Output $PrimaryKey
#            Write-Output $Type
#            Write-Output ===============<

            (Get-Content ${OUTFILE}) | ForEach-Object {
                $_ -replace '{{Res}}', $Res -replace '{{ConfigFile}}', $ConfigFile -replace '{{PrimaryKey}}', $PrimaryKey -replace '{{Type}}', $Type
            } | Set-Content ${OUTFILE}
        }
    }

}