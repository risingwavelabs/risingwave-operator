#!/bin/bash
set -e

# get commit_id
COMMIT_ID="main"
if [ $RISINGWAVE_DASHBOARD_COMMIT_ID ];then
	COMMIT_ID=$RISINGWAVE_DASHBOARD_COMMIT_ID
else
	echo "Environment variable \"RISINGWAVE_DASHBOARD_COMMIT_ID\" not found, use \"main\" as commit_id"
fi

shell_url="https://raw.githubusercontent.com/risingwavelabs/risingwave/${COMMIT_ID}/grafana/generate.sh"
python_url="https://raw.githubusercontent.com/risingwavelabs/risingwave/${COMMIT_ID}/grafana/risingwave-dashboard.dashboard.py"

# download ./generate.sh and risingwave-dashboard.dashboard.py
wget $shell_url -O "./generate.sh"
wget $python_url -O "risingwave-dashboard.dashboard.py"

chmod +x ./generate.sh

# generate risingwave-dashboard.json
DASHBOARD_NAMESPACE_FILTER_ENABLED=true DASHBOARD_RISINGWAVE_NAME_FILTER_ENABLED=true DASHBOARD_SOURCE_UID="prometheus" ./generate.sh

# replace for cloud deployment, will read risingwave-dashboard.json and write the result into risingwave-dashboard_new.json
python3 ./convert.py

# remove intermediate files
rm risingwave-dashboard.dashboard.py
rm risingwave-dashboard.gen.json
rm ./generate.sh
rm risingwave-dashboard.json

# rename
mv risingwave-dashboard_new.json risingwave-dashboard.json

echo "risingwave-dashboard.json updated"
