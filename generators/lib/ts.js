const {docBuilders: {hardline}, printDocToString, TsPrinter} = require("eapi");

function print(t, {getConfig}) {
  console.log('getConfig(name)', getConfig('name'))

  const doc = [];
  const printer = new TsPrinter(t);
  const keys = Object.keys(t.components.schemas).sort();
  keys.forEach(function (key) {
    const schema = t.components.schemas[key]
    if (schema.$ref) return;
    const ext = schema.ext;
    if (ext && ext.type === 'specific') return;
    doc.push(printer.typeDef(schema), hardline, hardline)
  })
  const code = printDocToString(doc, {printWidth: 80, tabWidth: 2}).formatted
  return [
    {
      fileName: 'types.ts',
      code: code,
    }
  ]
}

module.exports = {print}
