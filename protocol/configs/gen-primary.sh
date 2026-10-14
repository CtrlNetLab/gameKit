#!/bin/bash

set -o errexit -o nounset -o pipefail

PKG_DIR="${PKG_DIR:-../../pkg}"
CONFIGS_DIR="${CONFIGS_DIR:-./}"

gen_primary() {
    grep "Config:" "${CONFIGS_DIR}/primarykey.txt" | while IFS=':' read -r DataConfig PrimaryKey Type; do
        ProtoFile="${CONFIGS_DIR}/${DataConfig}.proto"
        test -f "${ProtoFile}" || continue

        Struct="$(sed -n 's#^message \([^{ \t]*\).*#\1#p' "${ProtoFile}" | head -n1)"
        # 首字母大写, 参考: https://www.shellscript.sh/examples/case/
        Res="${Struct^}"

        PbFile="${PKG_DIR}/configs/${Res}ResMgr.pb.go"

        sed -e 's#{{Res}}#'"${Res}"'#g' \
            -e 's#{{DataConfig}}#'"${DataConfig}"'#g' \
            -e 's#{{PrimaryKey}}#'"${PrimaryKey}"'#g' \
            -e 's#{{Type}}#'"${Type}"'#g' "${CONFIGS_DIR}/ResMgr.go.tmpl" >"${PbFile}"
    done
}

gen_primary "$@"
