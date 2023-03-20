const {hello} = require("./lib");

function print() {
  return [
    {
      fileName: 'custom-generator.ts',
      code: '// hello, ' + hello(),
    }
  ]
}

module.exports = {
  print,
}
