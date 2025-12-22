const fs = require('fs');
const path = require('path');
const inputJson = require('./tbosTransform.json');

function convertToJsFile(objectList) {
    objectList.forEach((item) => {
        // let sourceRes; let sourceReq; let targetReq; let targetRes;
        // try {
        //     sourceRes = item.targetReq ? JSON.parse(item.sourceRes) : item.targetReq;
        //     sourceReq = item.sourceReq ? JSON.parse(item.sourceReq) : item.sourceReq;
        //     targetReq = item.targetReq ? JSON.parse(item.targetReq) : item.targetReq;
        //     targetRes = item.targetRes ? JSON.parse(item.targetRes) : item.targetRes;
        // } catch (error) {
        // }
        const { targetPage } = item;
        const { apiName } = item;
        const { trueTargetPath } = item;
        const { sourcePath } = item;
        const { targetPath } = item;
        // const { checkTag } = item;
        // const { finishStatus } = item;
        // const { _id } = item;
        // const { updateTime } = item;
        // const { __v } = item;
        // const { createTime } = item;

        // 将当前对象存储在结果中，以 sourcePath 为键
        const result = {
            apiName,
            sourcePath,
            trueTargetPath,
            targetPage,
            targetPath,
            // _id,
            // updateTime,
            // sourceRes,
            // sourceReq,
            // __v,
            // createTime,
            // targetReq,
            // targetRes,
            // checkTag,
            // finishStatus,
        };
        let apiNameStr = ""
        apiNameStr = item?.apiName ? (item.apiName).replace('/', '') : '';
        const filePath = path.join('./funcs', `${path.basename(item.sourcePath)}-${apiNameStr}.js`);
        // 使用 replacer 函数移除对象键的双引号
        const content = `module.exports = (function() {
            const result = ${JSON.stringify(result, (key, value) => {
            if (typeof value === 'object' && value !== null) {
                const newObj = {};
                for (const k in value) {
                    newObj[k] = value[k];
                }
                return newObj;
            }
            return value;
        }, 2)};
            result.change = ${item.change};
            return result;
        })();`;
        fs.writeFileSync(filePath, content);
    });
}

// 示例使用
const objectList = inputJson;

// 确保输出目录存在
if (!fs.existsSync('./funcs')) {
    fs.mkdirSync('./funcs');
}
convertToJsFile(objectList);
