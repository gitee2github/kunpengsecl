/*
kunpengsecl licensed under the Mulan PSL v2.
You can use this software according to the terms and conditions of
the Mulan PSL v2. You may obtain a copy of Mulan PSL v2 at:
    http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

#include "ktalib.h"

#define OPERATION_START_FLAG 1
#define PARAMETER_FRIST 0
#define PARAMETER_SECOND 1
#define PARAMETER_THIRD 2
#define PARAMETER_FOURTH 3
#define RSA_PUB_SIZE 256

struct ktaprivkey {
    uint8_t modulus[RSA_PUB_SIZE];
    uint8_t privateExponent[RSA_PUB_SIZE];
};
static const TEEC_UUID Uuid = {
    0x435dcafa, 0x0029, 0x4d53, { 0x97, 0xe8, 0xa7, 0xa1, 0x3a, 0x80, 0xc8, 0x2e }
};
enum TEEC_Return{
    TEEC_ERROR_BAD_BUFFER_DATA = 0xFFFF0006
};
enum{
    INITIAL_CMD_NUM = 0x7FFFFFFF
};
enum {
    CMD_KTA_INITIALIZE      = 0x00000001, //send key and cert to kta for initialization, get kta public-key certificate
    CMD_SEND_TAHASH         = 0x00000002, //send ta hash values to kta for local attastation
    CMD_GET_REQUEST         = 0x00000003, //ask kta for commands in its cmdqueue
    CMD_RESPOND_REQUEST     = 0x00000004, //reply a command to kta(maybe one)
    CMD_CLOSE_KTA           = 0x00000006,
};

TEEC_Context context = {0};
TEEC_Session session = {0};

/*编译方法
    gcc -fPIC -shared -o libkta.so ktalib.c ./itrustee_sdk/src/CA/libteec_adaptor.c -I ./itrustee_sdk/include/CA/
*/

// 初始化上下文和会话
TEEC_Result InitContextSession(uint8_t* ktapath) {
    TEEC_Operation operation = {0};
    uint32_t origin = 0;
    TEEC_Result ret;

    ret = TEEC_InitializeContext(NULL, &context);
    if (ret != TEEC_SUCCESS) {
        return ret;
    }
    context.ta_path = ktapath;
    operation.started = OPERATION_START_FLAG;
    operation.paramTypes = TEEC_PARAM_TYPES(
        TEEC_NONE,
        TEEC_NONE,
        TEEC_NONE,
        TEEC_NONE);

    ret = TEEC_OpenSession(&context, &session, &Uuid, TEEC_LOGIN_IDENTIFY, NULL, &operation, &origin);
    if (ret != TEEC_SUCCESS) {
        TEEC_FinalizeContext(&context);
        return ret;
    }
    return ret;
}

// 向KTA发出初始化命令
TEEC_Result KTAinitialize(struct buffer_data* kcmPubKey_N, struct buffer_data* ktaPubCert, struct buffer_data* ktaPrivKey_N, struct buffer_data* ktaPrivKey_D, struct buffer_data *out_data){
    TEEC_Operation operation = {0};
    uint32_t origin = 0;
    TEEC_Result ret;
    struct ktaprivkey ktaPrivKey = {0};
    memcpy(ktaPrivKey.modulus, ktaPrivKey_N->buf, ktaPrivKey_N->size);
    memcpy(ktaPrivKey.privateExponent, ktaPrivKey_D->buf, ktaPrivKey_D->size);
    operation.started = OPERATION_START_FLAG;
    operation.paramTypes = TEEC_PARAM_TYPES(
        TEEC_MEMREF_TEMP_INPUT,   //存放KCM公钥
        TEEC_MEMREF_TEMP_INPUT,   //存放KTA公钥证书
        TEEC_MEMREF_TEMP_INPUT,   //存放KTA私钥
        TEEC_MEMREF_TEMP_OUTPUT  //存放KTA公钥证书（返回）
    );

    operation.params[PARAMETER_FRIST].tmpref.buffer = kcmPubKey_N->buf;
    operation.params[PARAMETER_FRIST].tmpref.size = kcmPubKey_N->size;
    operation.params[PARAMETER_SECOND].tmpref.buffer = ktaPubCert->buf;
    operation.params[PARAMETER_SECOND].tmpref.size = ktaPubCert->size;
    operation.params[PARAMETER_THIRD].tmpref.buffer = &ktaPrivKey;
    operation.params[PARAMETER_THIRD].tmpref.size = sizeof(struct ktaprivkey);
    operation.params[PARAMETER_FOURTH].tmpref.buffer = out_data->buf;
    operation.params[PARAMETER_FOURTH].tmpref.size = out_data->size;

    ret = TEEC_InvokeCommand(&session, CMD_KTA_INITIALIZE, &operation, &origin);
    if (ret != TEEC_SUCCESS) {
        return ret;
    }

    return TEEC_SUCCESS;
}


// 向KTA发送TA哈希
TEEC_Result KTAsendHash(struct buffer_data* in_data, uint32_t innum) {
    TEEC_Operation operation = {0};
    uint32_t origin = 0;
    TEEC_Result ret;
    operation.started = OPERATION_START_FLAG;
    operation.paramTypes = TEEC_PARAM_TYPES(
        TEEC_MEMREF_TEMP_INPUT, //存放传入KTA的哈希
        TEEC_NONE,
        TEEC_NONE,
        TEEC_VALUE_INPUT //存放KA发送的哈希数量：0<a<=32
    );
    operation.params[PARAMETER_FRIST].tmpref.buffer = in_data->buf;
    operation.params[PARAMETER_FRIST].tmpref.size = in_data->size;
    operation.params[PARAMETER_FOURTH].value.a = innum;
    ret = TEEC_InvokeCommand(&session, CMD_SEND_TAHASH, &operation, &origin);
    if (ret != TEEC_SUCCESS) {
        return ret;
    }
    return TEEC_SUCCESS;
}

// 从KTA拿取密钥请求
TEEC_Result KTAgetCommand(struct buffer_data* out_data, uint32_t* retnum){
    TEEC_Operation operation = {0};
    uint32_t origin = 0;
    TEEC_Result ret;
    operation.started = OPERATION_START_FLAG;
    operation.paramTypes = TEEC_PARAM_TYPES(
        TEEC_MEMREF_TEMP_OUTPUT, //存放请求
        TEEC_VALUE_OUTPUT, //存放剩余请求数量(包含此次得到的的请求在内)
        TEEC_NONE,
        TEEC_NONE
    );
    operation.params[PARAMETER_FRIST].tmpref.buffer = out_data->buf;
    operation.params[PARAMETER_FRIST].tmpref.size = out_data->size;
    operation.params[PARAMETER_SECOND].value.a = INITIAL_CMD_NUM;
    operation.params[PARAMETER_SECOND].value.b = INITIAL_CMD_NUM;
    ret = TEEC_InvokeCommand(&session, CMD_GET_REQUEST, &operation, &origin);
    if (ret != TEEC_SUCCESS) {
        return ret;
    }
    *retnum = operation.params[PARAMETER_SECOND].value.a;
    out_data->size = operation.params[PARAMETER_SECOND].value.b;
    return TEEC_SUCCESS;
}

// 向KTA返回密钥请求结果
TEEC_Result KTAsendCommandreply(struct buffer_data* in_data){
    TEEC_Operation operation = {0};
    uint32_t origin = 0;
    TEEC_Result ret;

    operation.started = OPERATION_START_FLAG;
    operation.paramTypes = TEEC_PARAM_TYPES(
        TEEC_MEMREF_TEMP_INPUT, //存放请求结果
        TEEC_VALUE_OUTPUT, //存放KTA处理结果
        TEEC_NONE,
        TEEC_NONE
    );
    operation.params[PARAMETER_FRIST].tmpref.buffer = in_data->buf;
    operation.params[PARAMETER_FRIST].tmpref.size = in_data->size;
    operation.params[PARAMETER_SECOND].value.a = INITIAL_CMD_NUM;
    ret = TEEC_InvokeCommand(&session, CMD_RESPOND_REQUEST, &operation, &origin);
    if (ret != TEEC_SUCCESS) {
        return ret;
    }
    if (operation.params[PARAMETER_SECOND].value.a == 0){
        return TEEC_ERROR_BAD_BUFFER_DATA;
    }
    return TEEC_SUCCESS;
}

// 关闭与KTA的连接
void KTAshutdown() {
    TEEC_CloseSession(&session);
    TEEC_FinalizeContext(&context);
}

// 终止kta, 测试用
TEEC_Result KTAterminate(){
    TEEC_Operation operation = {0};
    uint32_t origin = 0;
    TEEC_Result ret;

    operation.started = OPERATION_START_FLAG;
    operation.paramTypes = TEEC_PARAM_TYPES(
        TEEC_NONE, 
        TEEC_NONE,
        TEEC_NONE,
        TEEC_NONE
    );
    ret = TEEC_InvokeCommand(&session, CMD_CLOSE_KTA, &operation, &origin);
    if (ret != TEEC_SUCCESS) {
        return ret;
    }
    return TEEC_SUCCESS;
}