module.exports = {
  development: {
    // 所有项目共用
    _common_: {
      header: 'dev-header.html',
      footer: 'dev-footer.html',
    },
    tnebula: {
      _common_: {
        header: 'header.html',
        body: 'body.html',
        footer: 'footer.html',
      },
    },
    tassets: {
      _common_: {
        header: 'dev-tnebula-header.html',
      },
    },
    tompage: {
      _common_: {
        header: 'dev-tnebula-header.html',
      },
    },
    tedge: {
      _common_: {
        header: 'dev-adaptor-header.html',
        footer: 'dev-adaptor-footer.html',
      },
    },
    monitor: {
      _common_: {
        header: 'dev-adaptor-header.html',
        footer: 'dev-adaptor-footer.html',
      },
    },
    tshows: {
      _common_: {
        header: 'dev-shows-header.html',
        footer: 'dev-shows-footer.html',
      },
    },
    idcdb: {
      _common_: {
        header: 'dev-tnebula-header.html',
      },
    },
  },
  production: {
    _common_: { header: 'pub-header.html' },
    tnebula: {
      _common_: {
        header: 'header.html',
        body: 'body.html',
        footer: 'footer.html',
      },
    },
    tompage: {
      _common_: {
        header: 'dev-tnebula-header.html',
      },
    },
    idcdb: {
      _common_: {
        header: 'dev-tnebula-header.html',
      },
    },
  },
};
