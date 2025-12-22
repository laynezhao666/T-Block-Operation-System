const cheerio = require('cheerio');
const LastCallWebpackPlugin = require('last-call-webpack-plugin');
const { getConfItem } = require('./config');

class PhpPlugin {
  constructor() {
    this.lastCallInstance = new LastCallWebpackPlugin({
      assetProcessors: [
        {
          phase: LastCallWebpackPlugin.PHASE.EMIT,
          regExp: /\.html$/,
          processor: (assetName, asset, assets) => this.process(assetName, asset, assets),
        },
      ],
    });
  }

  async process(assetName, asset, assets) {
    const chunkName = (assetName.split('.')[0]).split('/');
    const filePathName = chunkName.pop();
    const basePath = `${chunkName.join('/')}/${getConfItem('assetsSubDirectory')}/${getConfItem('bundleDir')}/${filePathName}`;
    const jsPath = `${basePath}/js.php`;
    const cssPath = `${basePath}/css.php`;
    const $ = cheerio.load(asset.source());
    assets.setAsset(jsPath, $('script').toString());
    assets.setAsset(cssPath, $('link').toString());
  }

  apply(compiler) {
    return this.lastCallInstance.apply(compiler);
  }
}

module.exports = PhpPlugin;
