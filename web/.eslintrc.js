const noDupeClassMember = require('./eslint/rules/no-dupe-class-members')

module.exports = {
  "env": {
    "browser": true,
    "es6": true
  },
  "extends": [
    "standard",
    "plugin:vue/recommended",
    // "@tencent/eslint-config-tencent"
  ],
  "plugins": [
    "html",
    "vue",
    "babel",
    "bug-overrides"
  ],
  "parser": "vue-eslint-parser",
  "parserOptions": {
    "parser": "babel-eslint"
  },
  "overrides": [
    {
      "files": [
        "*.ts.vue"
      ],
      "parser": "vue-eslint-parser",
      "parserOptions": {
        "parser": "typescript-eslint-parser"
      }
    },
    {
      "files": [
        "*.ts"
      ],
      "parser": "typescript-eslint-parser"
    }
  ],
  "rules": {
    // "indent": "off",
    // "template-curly-spacing": "off",
    "no-extend-native": 0,
    "babel/no-unused-expressions": 1,
    "no-dupe-class-members": 0,
    "bug-overrides/no-dupe-class-members": 2
  },
  "globals": {
    "openDialog": true,
    "echarts": true,
    "_": true,
    "$": true,
    "FO": true,
    "Global": true,
    "HOST": true,
    "HOME": true,
    "TNBL": true,
    "ht": true
  }
}
