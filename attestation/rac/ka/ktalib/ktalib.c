#include "ktalib.h"

#define OPERATION_START_FLAG 1
#define MAX_STR_LEN 200 //now set randomly
#define PARAMETER_FRIST 0
#define PARAMETER_SECOND 1
#define PARAMETER_THIRD 2
#define PARAMETER_FOURTH 3
#define SIGEND_PUBKEY_BUF 1000 //now set randomly
#define TEE_PUBKEY_BUF 500 //now set randomly
#define KCM_ENCRYPT_BUF 600 //now set randomly
#define KMS_PUBKEY_BUF 800 //now set randomly

static const TEEC_UUID Uuid = {
    0x435dcafa, 0x0029, 0x4d53, { 0x97, 0xe8, 0xa7, 0xa1, 0x3a, 0x80, 0xc8, 0x2e }
};
enum TEEC_Return{
    TEEC_ERROR_BAD_BUFFER_DATA = 0xFFFF0006
};

enum {
    CMD_KTA_INITIALIZE      = 0x00000001, //send request to kta for setup snd initialization, get parameters kta generated during initialization
    CMD_KTA_INITREPLY       = 0x00000002, //send reply to kta for initialization
    CMD_GET_REQUEST         = 0x00000003, //ask kta for commands in its cmdqueue, and send ta identification whose trusted status needs to update
    CMD_RESPOND_REQUEST     = 0x00000004, //reply a command to kta(maybe one)
};

TEEC_Context initcontext(TEEC_Context context) {
    context.ta_path = "/root/data/435dcafa-0029-4d53-97e8-a7a13a80c82e.sec"; //to be set, the path of kta mirror
    return context;
}
TEEC_Result initialize(TEEC_Context *context, TEEC_Session *session, struct buffer_data* pubKey,struct buffer_data* pubCert){
    TEEC_Operation operation = {0};
    uint32_t origin = 0;
    TEEC_Result ret;
    TEEC_Value initresult = {0};
    char ktaSignedPubkey[SIGEND_PUBKEY_BUF] = {0};
    char teePubkey[TEE_PUBKEY_BUF] = {0};
    char kcmEncryptKey[KCM_ENCRYPT_BUF] = {0};
    char kmsPubKey[KMS_PUBKEY_BUF] = {0};
    unsigned int pubkeyBufLen = SIGEND_PUBKEY_BUF;
    unsigned int teePubkeyBufLen = TEE_PUBKEY_BUF;
    unsigned int kcmKeyBufLen = KCM_ENCRYPT_BUF;
    unsigned int kmsPubkeyBufLen = KMS_PUBKEY_BUF;

    ret = TEEC_InitializeContext(NULL, context);
    if (ret != TEEC_SUCCESS) {
        tloge("init context is failed! result is 0x%x\n", ret);
        return TEEC_ERROR_GENERIC;
    }
    operation.started = OPERATION_START_FLAG;
    operation.paramTypes = TEEC_PARAM_TYPES(
        TEEC_NONE,
        TEEC_NONE,
        TEEC_NONE,
        TEEC_NONE);

    ret = TEEC_OpenSession(context, session, &Uuid, TEEC_LOGIN_IDENTIFY, NULL, &operation, &origin);
    if (ret != TEEC_SUCCESS) {
        tloge("open session is failed! result is 0x%x\n", ret);
        TEEC_FinalizeContext(context);
        return 0;
    }

    operation.started = OPERATION_START_FLAG;
    operation.paramTypes = TEEC_PARAM_TYPES(
        TEEC_VALUE_OUTPUT, //存放初始化结果，具体有待商议
        TEEC_MEMREF_TEMP_OUTPUT, //本来是存放TEE证书私钥签名的KTA公钥，现在待定
        TEEC_MEMREF_TEMP_OUTPUT, //本来是存放TEE设备公钥，现在待定
        TEEC_NONE
    );
    //将KCM的公钥作为operation的参数，调用InvokeCommand函数，由operation接收到KTA的公钥证书

    operation.params[PARAMETER_FRIST].value = initresult;
    operation.params[PARAMETER_SECOND].tmpref.buffer = pubKey->buf;
    operation.params[PARAMETER_SECOND].tmpref.size = pubKey->size;

    ret = TEEC_InvokeCommand(session, CMD_KTA_INITIALIZE, &operation, &origin);
    if (ret != TEEC_SUCCESS) {
        tloge("kta initialize failed, codes=0x%x, origin=0x%x", ret, origin);
        return ret;
    } else if (initresult.a != 0x01) {
        /* 到kta初始化失败处理逻辑 */
    }
    //从operation中获取到KTA的公钥证书
    pubCert->buf = operation.params[PARAMETER_SECOND].tmpref.buffer; //在operation的第二个参数中存放有返回来的kta的公钥证书
    pubCert->size = operation.params[PARAMETER_SECOND].tmpref.size;

    return TEEC_SUCCESS;
}

//request is the kcm's public key and the response is the kta's cert
TEEC_Result RemoteAttestInitial(uint32_t cmdnum,struct buffer_data *req,struct buffer_data*rsp){
    TEEC_Result ret = TEEC_SUCCESS;
    if(req==NULL||req->buf==NULL||req->size<0){
        return TEEC_ERROR_SECURITY;
    }
    return ret;
}
TEEC_Result RemoteAttestKTA(uint32_t cmdnum,struct buffer_data *req,struct buffer_data *rsp)
{   //根据cmdnum进行相关处理
    if(req==NULL||req->buf==NULL||rsp==NULL||rsp->size==0){
        //tloge("bad input request or short out data size\n");
        return TEEC_ERROR_BAD_BUFFER_DATA;
    }
    TEEC_Result ret = TEEC_SUCCESS;
    struct buffer_data out_data; 
    TEEC_Context context = {0};
    TEEC_Session session = {0};
    //init kta
    context = initcontext(context);
    ret = initialize(&context,&session,req,&out_data);
    if (ret != TEEC_SUCCESS) {
        printf("kta initialize failed, codes=0x%x", ret);
        return ret;
    }
    //get kta's request by operation
    printf("%d\n",req->size);
    
    return TEEC_SUCCESS;
}
/*编译方法
    gcc -fPIC -shared -o libkta.so ktalib.c ./itrustee_sdk/src/CA/libteec_adaptor.c -I ./itrustee_sdk/include/CA/
*/
// int main(){
//     return 0;
// }