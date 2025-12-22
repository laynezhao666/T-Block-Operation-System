const noDupeClassMembers = require('./rules/no-dupe-class-members');

// use commonjs default export so ESLint can find the rule
module.exports = {
  rules: {
    'no-dupe-class-members': noDupeClassMembers,
  },
};
