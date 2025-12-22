
module.exports = function (api) {
  api.cache(true);
  return {
    presets: [
      ['@babel/preset-env', {
        modules: false,
        targets: {
          browsers: ['> 1%', 'last 2 versions', 'not ie <= 8'],
        },
      },
      ],
    ],
    plugins: [
      '@babel/plugin-proposal-optional-chaining',
      'transform-vue-jsx',
      '@babel/plugin-transform-runtime',
      '@babel/plugin-proposal-export-default-from',
      ['@babel/plugin-proposal-pipeline-operator', { proposal: 'minimal' }],
      '@babel/plugin-proposal-nullish-coalescing-operator',
      ['@babel/plugin-proposal-decorators', { decoratorsBeforeExport: true }],
      '@babel/plugin-proposal-export-namespace-from',
      '@babel/plugin-syntax-dynamic-import',
      '@babel/plugin-proposal-class-properties',
      '@babel/plugin-proposal-private-methods',
      '@babel/plugin-proposal-private-property-in-object',
    ],
    ignore: [
      '**/ht/**',
    ],
  };
};
