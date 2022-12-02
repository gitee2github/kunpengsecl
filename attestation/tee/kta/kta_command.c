/*
kunpengsecl licensed under the Mulan PSL v2.
You can use this software according to the terms and conditions of
the Mulan PSL v2. You may obtain a copy of Mulan PSL v2 at:
    http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.

Author: leezhenxiang
Create: 2022-11-04
Description: api module in kta.
	1. 2022-11-04	leezhenxiang
		define the structures.
    2. 2022-11-18   waterh2o
        redefine some interface
*/

#include <tee_defines.h>
#include <kta_command.h>

#define MAX_CERT_LEN 8192
#define MAX_KEY_LEN 2048
#define PARAM_COUNT 4

TEE_Result KTAInitialize(uint32_t param_type, TEE_Param params[PARAM_COUNT],Cache *cache,CmdQueue *cmdqueue){
    //basic function for calling the above functions
    TEE_Result ret;
    if (!check_param_type(param_type,
        TEE_PARAM_TYPE_MEMREF_INPUT,
        TEE_PARAM_TYPE_MEMREF_INPUT,
        TEE_PARAM_TYPE_MEMREF_INPUT,
        TEE_PARAM_TYPE_MEMREF_OUTPUT)) {
        tloge("Bad expected parameter types, 0x%x.\n", param_type);
        return TEE_ERROR_BAD_PARAMETERS;
    }
    if (params[0].memref.size == 0 || params[0].memref.size > MAX_KEY_LEN || params[0].memref.buffer == NULL) {
        tloge("Bad expected parameter.\n");
        return TEE_ERROR_BAD_PARAMETERS;
    }
    if (params[1].memref.size == 0 || params[1].memref.size > MAX_CERT_LEN || params[1].memref.buffer == NULL) {
        tloge("Bad expected parameter.\n");
        return TEE_ERROR_BAD_PARAMETERS;
    }
    if (params[2].memref.size == 0 || params[2].memref.size > MAX_KEY_LEN || params[2].memref.buffer == NULL) {
        tloge("Bad expected parameter.\n");
        return TEE_ERROR_BAD_PARAMETERS;
    }
    if (params[3].memref.size == 0 || params[3].memref.buffer == NULL) {
        tloge("Bad expected parameter.\n");
        return TEE_ERROR_BAD_PARAMETERS;
    }
    ret = saveKeyPair("sec_storage_data/kcmpub.txt", params[0].memref.buffer, params[0].memref.size, TEE_TYPE_RSA_PUBLIC_KEY);
    if (ret != TEE_SUCCESS){
        tloge("save kcmpub failed\n");
        return ret;
    }
    ret = saveCert("sec_storage_data/ktacert.txt", params[1].memref.buffer, params[1].memref.size);
    if (ret != TEE_SUCCESS){
        tloge("save kta cert failed\n");
        return ret;
    }
    ret = saveKeyPair("sec_storage_data/ktakey.txt", params[2].memref.buffer, params[2].memref.size, TEE_TYPE_RSA_KEYPAIR);
    if (ret != TEE_SUCCESS){
        tloge("save ktakey failed\n");
        return ret;
    }
    ret = initStructure(cache,cmdqueue);
    if (ret != TEE_SUCCESS){
        tloge("init kta struct failed\n");
        return ret;
    }

    ret = restoreCert("sec_storage_data/ktacert.txt",params[3].memref.buffer, &params[3].memref.size);
    if (ret != TEE_SUCCESS){
        tloge("restore kta cert failed\n");
        return ret;
    }
    return TEE_SUCCESS;
}

TEE_Result SendRequest() {
    //todo: send request to ka when ka polls, and answer ta trusted state which ka asks
}

TEE_Result GetResponse() {
    //todo: Get Response from ka when kta had sent request to kcm before
}

// Communication with ta

TEE_Result InitTAKey(TEE_UUID TA_uuid, Cache *cache) {
    //todo: init ta cache;
};

TEE_Result SendReplytoTA() {
    //todo: answer to ta when ta asks its command's reply.
}

//the following operation must start with ta authenticating

TEE_Result SearchTAKey(TEE_UUID TA_uuid, char *keyid, Cache *cache, char *keyvalue) {
    //todo: search a certain ta key, if not exist, call AddTATable() to add a request

    //input: TA_uuid, keyid, cache
    //output: cache, keyvalue
}

TEE_Result DeleteTAKey(TEE_UUID TA_uuid, char *keyid, Cache *cache) {
    //todo: delete a certain key in the cache

    //input: TA_uuid, keyid, cache
    //output: cache
}

TEE_Result DestoryKey(TEE_UUID TA_uuid, char *keyid, Cache *cache) {
    //todo: delete a certain key by calling DeleteTAKey(), then generate a delete key request in TaCache

    //input: TA_uuid, keyid, cache
    //output: cache
}
