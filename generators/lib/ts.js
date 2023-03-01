const {docBuilders: {hardline, join}, printDocToString, TsPrinter} = require("eapi");

function print(t, {getConfig}) {
  const types = [];
  const printer = new TsPrinter(t);
  const keys = Object.keys(t.components.schemas).sort();
  keys.forEach(function (key) {
    const schema = t.components.schemas[key]
    if (schema.$ref) return;
    const ext = schema.ext;
    if (ext && ext.type === 'specific') return;
    types.push(printer.typeDef(schema))
  })
  const doc = [join([hardline, hardline], types), hardline];
  const code = printDocToString(doc, {printWidth: 80, tabWidth: 2}).formatted
  return [
    {
      fileName: 'types.ts',
      code: code,
    }
  ]
}

module.exports = {print}
