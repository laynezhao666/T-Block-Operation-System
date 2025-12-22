const fs = require('fs');
const path = require('path');

function readFilesAndConvertToJson(folderPath, outputFilePath) {
    const resultList = [];
    // 将 folderPath 转换为绝对路径
    const absoluteFolderPath = path.resolve(folderPath);
    if (!fs.existsSync(absoluteFolderPath)) {
        console.error(`The folder ${absoluteFolderPath} does not exist.`);
        return;
    }
    const files = fs.readdirSync(absoluteFolderPath);
    files.forEach((file) => {
        const filePath = path.join(absoluteFolderPath, file);
        if (fs.statSync(filePath).isFile()) {
            try {
                const object = require(filePath);
                // 将 change 函数转换为字符串，保留函数名
                // const changeFunctionString = object.change.toString();
                const changeFunction = object.change.toString();
                const objectWithChangeAsString = {
                    ...object,
                    change: changeFunction,
                };
                resultList.push(objectWithChangeAsString);
            } catch (error) {
                console.error(`Error reading file ${filePath}:`, error);
            }
        }
    });
    fs.writeFileSync(outputFilePath, JSON.stringify(resultList, null, 2));
}

// 示例使用
const folderPath = './funcs'; // 相对路径
const outputFilePath = './transformMap.json';
readFilesAndConvertToJson(folderPath, outputFilePath);
