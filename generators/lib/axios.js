const {
  docBuilders: {hardline, softline, join, indent, group, ifBreak, indentIfBreak},
  printDocToString,
  TsPrinter,
  camelCase,
} = require("eapi");

const MIME_TYPE_FORM_DATA = "multipart/form-data"

const contentTypeInPriority = ["application/json", MIME_TYPE_FORM_DATA]

class Printer {
  /**
   * @type {OpenAPI}
   */
  openAPI = null;
  tsPrinter = null;
  options = null;

  constructor(openAPI, options) {
    this.openAPI = openAPI;
    this.options = options;
    this.tsPrinter = new TsPrinter(openAPI);
  }

  print() {
    const requests = [];
    Object.keys(this.openAPI.paths).sort().forEach(path => {
      const pathItem = this.openAPI.paths[path]
      requests.push(...this.pathItem(path, pathItem))
    })
    const header = this.options.getConfig('customHeader') || 'import axios, { AxiosRequestConfig } from "axios";';

    const imports = group([
      header, hardline,
      'import {',
      indent([
        hardline,
        join([',', hardline], this.tsPrinter.importedTypes),
      ]), hardline,
      '} from \"./types\";',
    ]);
    return [
      imports, hardline, hardline,
      join([hardline, hardline], requests),
    ]
  }

  /**
   * @param {string} path
   * @param {PathItem}  item
   */
  pathItem(path, item) {
    const res = [];
    if (item.connect) res.push(this.request(path, 'connect', item.connect))
    if (item.delete) res.push(this.request(path, 'delete', item.delete))
    if (item.get) res.push(this.request(path, 'get', item.get))
    if (item.head) res.push(this.request(path, 'head', item.head))
    if (item.options) res.push(this.request(path, 'options', item.options))
    if (item.patch) res.push(this.request(path, 'patch', item.patch))
    if (item.post) res.push(this.request(path, 'post', item.post))
    if (item.put) res.push(this.request(path, 'put', item.put))
    if (item.trace) res.push(this.request(path, 'trace', item.trace))
    return res;
  }

  request(path, method, operation) {
    const fnName = this.fnName(operation.operationId);
    const args = [];
    let pathParams = [], queryParams = [];
    operation.parameters.forEach(p => {
      switch (p.in) {
        case 'path':
          pathParams.push(p)
          break
        case 'query':
          queryParams.push(p)
          break
      }
    })

    // try to parse path params from path
    let pathName = path.replaceAll("{", "${")
    const pathArgMatches = pathName.matchAll(/\{([\w-]+)\}/g)
    for (let match of pathArgMatches) {
      const originalName = match[1];
      const argName = camelCase(originalName);
      pathName = pathName.replaceAll(`{${originalName}`, `{${argName}`)
      if (originalName !== argName) {
        pathParams = pathParams.filter(p => p.name !== originalName)
      }
      if (!pathParams.find(p => p.name === argName)) {
        pathParams.push({name: argName, in: 'path', required: true})
      }
    }
    pathParams.forEach(p => {
      args.push([p.name, ': string'])
    })

    const funcBody = [];
    const options = [];
    if (queryParams.length > 0) {
      options.push(['params: query,'])
      args.push(['query: ', '{ ', join('; ', queryParams.map(p => [p.name, p.required ? ': ' : '?: ', this.tsPrinter.typeName(p.schema)])), ' }'])
    }

    if (operation.requestBody) {
      const {contentType, data} = this.requestBody(operation.requestBody) || {};
      if (data) {
        args.push(['data: ', data])
        if (contentType === MIME_TYPE_FORM_DATA) {
          funcBody.push([
            'const formData = new FormData();', hardline,
            'Object.keys(data).forEach((key) => formData.append(key, data[key]));', hardline,
          ])
          options.push(['data: formData,'])
        } else {
          options.push(['data,'])
        }
      }
    }

    args.push(['config?: AxiosRequestConfig'])
    options.push(['...config,'])
    return [
      this.apiDoc(operation),
      'export function ', fnName, '(', join(', ', args), ') {',
      indent([
        hardline,
        [
          funcBody,
          `return axios.`, method, this.responseType(operation.responses), `(\`${pathName}\`, {`,
          indent([
            hardline,
            join(hardline, options),
          ]), hardline,
          '});'
        ]
      ]), hardline,
      '}',
    ];
  }

  requestBody(requestBody) {
    if (!requestBody || !requestBody.content) {
      return null
    }
    for (const contentType of contentTypeInPriority) {
      const data = requestBody.content[contentType]
      if (data) {
        return {
          contentType: contentType,
          data: this.tsPrinter.typeName(data.schema)
        }
      }
    }
    return null;
  }

  responseType(responses) {
    for (let status = 200; status < 300; status++) {
      const response = responses[`${status}`];
      if (response) {
        let schema = null;
        for (const key of Object.keys(response.content)) {
          schema = response.content[key].schema;
        }
        return ['<', schema ? this.tsPrinter.typeName(schema) : 'any', '>']
      }
    }
    return [];
  }

  apiDoc(operation) {
    if (!operation.description.length) {
      return [];
    }
    const doc = this.tsPrinter.comment(operation.description.split("\n\n").map(
      (line, idx) => {
        if (idx === 0) return ['@description ', line]
        return ["\t", line]
      })
    )
    doc.push(hardline)
    return doc;
  }

  fnName(operationId) {
    return camelCase(operationId, {pascalCase: false})
  }
}

const tsPrinter = require("eapi/generators/ts")

/**
 * @param {OpenAPI} t
 * @returns {*}
 */
function print(t, options) {
  const printer = new Printer(t, options);
  const doc = printer.print()
  const code = printDocToString(doc, {printWidth: 80, tabWidth: 2}).formatted

  return [
    {
      fileName: 'request.ts',
      code: code,
    },
    ...tsPrinter.print(t, options),
  ]
}

module.exports = {print}
