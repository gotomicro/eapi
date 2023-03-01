import {group, hardline, indent, join} from "./doc-builders"

const SCHEMAS_REF_PREFIX = '#/components/schemas/';

class Printer {
  /**
   * @type {OpenAPI}
   */
  doc = null;

  /**
   * @type {string[]}
   */
  importedTypes = [];

  constructor(doc) {
    this.doc = doc
  }

  /**
   * @param {Schema} schema
   */
  typeDef(schema) {
    const doc = this.schemaDoc(schema);
    const ext = schema.ext;
    if (ext && ext.type === 'enum') {
      return [
        doc,
        'export enum ', schema.title, ' {',
        indent([
          hardline,
          join(
            hardline,
            ext.enumItems.map(e => {
              let value = e.value;
              if (typeof value === 'string') value = `"${value}"`
              return [e.key, ' = ', value + '', ',']
            })
          )
        ]), hardline,
        '}'
      ]
    }

    return group([
      doc,
      "export type ",
      schema.title,
      this.typeParams(schema),
      " = ",
      this.typeBody(schema)
    ])
  }

  schemaDoc(schema) {
    if (!schema.description.length) return []

    return [
      this.comment(schema.description.split("\n\n").map(
        (line, idx) => {
          if (idx === 0) return ['@description ', line]
          return ["\t", line]
        })
      ),
      hardline,
    ]
  }

  /**
   * @param {string[]} lines
   */
  comment(lines) {
    return [
      "/*", hardline,
      join(hardline, lines.map(l => [' * ', l])), hardline,
      " */",
    ]
  }

  /**
   * @param {Schema} schema
   */
  typeBody(schema) {
    const ref = schema.$ref;
    if (ref) {
      const typeName = this.unRef(ref)?.title;
      this.pushImportTypes(typeName)
      return typeName || 'unknown';
    }

    const ext = schema.ext;
    switch (ext?.type) {
      case "any":
        return "any";
      case 'array':
        if (!schema.items) return 'any[]';
        return [this.typeName(schema.items), '[]']
      case "map":
        return ['Record', '<', this.typeName(ext.mapKey), ', ', this.typeName(ext.mapValue), '>']
      case 'object':
        return this.printObjectBody(schema)
      case 'specific':
        return this.printSpecificType(ext)
      case 'param':
        return ext.typeParam.name;
      case 'null':
        return 'null'
      case 'unknown':
        return 'unknown'
    }

    return this.basicType(schema.type);
  }

  /**
   * @param {Schema} schema
   */
  typeName(schema) {
    const ref = schema.ref;
    if (ref) {
      schema = this.unRef(ref);
      switch (schema?.ext?.type) {
        case 'specific':
          return this.typeName(schema)
      }
      const typeName = schema?.title;
      this.pushImportTypes(typeName)
      return typeName || 'unknown';
    }

    return this.typeBody(schema);
  }

  /**
   * @param {string} ref
   */
  unRef(ref) {
    if (!ref || !ref.startsWith(SCHEMAS_REF_PREFIX)) {
      return undefined;
    }
    return this.doc.components.schemas[ref.substring(SCHEMAS_REF_PREFIX.length)]
  }

  pushImportTypes(typeName) {
    this.importedTypes = [...new Set([...this.importedTypes, typeName])]
  }

  /**
   * @param {Schema} schema
   */
  typeParams(schema) {
    const ext = schema.ext;
    if (!ext || !ext.typeParams?.length) {
      return []
    }
    return ['<', join(", ", ext.typeParams.map(p => [p.name, this.genericParamConstraint(p.constraint)])), '>'];
  }

  genericParamConstraint(constraint) {
    switch (constraint) {
      case 'comparable':
        return [' extends ', 'string | number']
    }

    return []
  }

  basicType(t) {
    const res = {
      "string": "string",
      "number": "number",
      "integer": "number",
      "boolean": "boolean",
      "file": "File",
    }[t];
    if (res) return res;
    console.log('unknown type', 'schema.type=' + t.type, 'ext.type=' + t.ext?.type)
    return 'unknown'
  }

  printObjectBody(schema) {
    const properties = schema.properties;
    return group([
      "{",
      indent([
        hardline,
        join(
          hardline,
          Object.keys(properties).filter(k => !!properties[k]).sort().map(k => {
            const p = properties[k]
            const required = schema.required?.includes(k);
            return [
              this.schemaDoc(p),
              k, required ? ':' : '?:', ' ', this.typeName(p), ';'
            ]
          }))
      ]), hardline,
      "}"
    ]);
  }

  printSpecificType(ext) {
    return [
      this.typeName(ext.specificType.type), '<', join(', ', ext.specificType.args.map(t => this.typeName(t))), '>',
    ]
  }
}

export default Printer;
