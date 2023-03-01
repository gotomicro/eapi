/**
 * @typedef Schema
 * @type {object}
 * @property {string} ref
 * @property {string} title
 * @property {string} type
 * @property {Object.<string, Schema>} properties
 * @property {Ext} ext
 * @property {Schema} items
 * @property {string[]} required
 *
 * @typedef {Object.<string, PathItem>} Paths
 *
 * @typedef {object} PathItem
 * @property {Operation} connect
 * @property {Operation} delete
 * @property {Operation} get
 * @property {Operation} head
 * @property {Operation} options
 * @property {Operation} patch
 * @property {Operation} post
 * @property {Operation} put
 * @property {Operation} trace
 *
 * @typedef {object} Operation
 * @property {string} summary
 * @property {string} description
 * @property {Parameter[]} parameters
 *
 * @typedef {object} Parameter
 * @property {string} ref
 *
 * @typedef Ext
 * @type {object}
 *
 * @typedef Components
 * @type {Object.<string, Schema>}
 *
 * @typedef OpenAPI
 * @type {object}
 * @property {Components} components
 * @property {Paths} paths
 */
