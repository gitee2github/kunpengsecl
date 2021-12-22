#!/bin/bash
# this scripts should be run under the root folder of kunpengsecl project
#set -eux
PROJROOT=.
# include common part
. ${PROJROOT}/attestation/test/integration/common.sh

###use integritytools to enable pcie-measurment
cd ${PROJROOT}/attestation/quick-scripts/integritytools
echo "enable host measurement" | tee -a ${DST}/control.txt
echo n | sh hostintegritytool.sh | tee -a ${DST}/control.txt
echo "enable pcie measurement" | tee -a ${DST}/control.txt
echo y | sh pcieintegritytool.sh | tee -a ${DST}/control.txt