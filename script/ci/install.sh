#!/bin/sh
set -e # Fail on errors
echo "installing libs..."
cd ..

# ----------------------------
# |     Library Versions     |
# ----------------------------
export GO_GAE_SDK=1.9.27


# ----------------------------
# |    Library Installers    |
# ----------------------------

# Load GO GAE SDK (include update check)
export CHECK=$(curl -s https://appengine.google.com/api/updatecheck | grep release | grep -o '"[0-9\.]*"')
if [ "$CHECK" != "\"$GO_GAE_SDK\"" ]; then
  echo "WARNING: New version of GO GAE SDK available" $CHECK
fi
echo "installing GO GAE SDK" $GO_GAE_SDK
curl -O https://storage.googleapis.com/appengine-sdks/featured/go_appengine_sdk_linux_amd64-$GO_GAE_SDK.zip
unzip -q go_appengine_sdk_linux_amd64-$GO_GAE_SDK.zip


echo "installing libs fetched."
exit 0
