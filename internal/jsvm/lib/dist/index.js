'use strict';

/**
 * TBD properly tagged union for Doc object type is needed here.
 *
 * @typedef {object} DocObject
 * @property {string} type
 * @property {boolean} [hard]
 * @property {boolean} [literal]
 *
 * @typedef {Doc[]} DocArray
 *
 * @typedef {string | DocObject | DocArray} Doc
 */

/**
 * @param {Doc} val
 */
function assertDoc(val) {
  if (typeof val === "string") {
    return;
  }

  if (Array.isArray(val)) {
    for (const doc of val) {
      assertDoc(doc);
    }
    return;
  }

  if (val && typeof val.type === "string") {
    return;
  }

  /* istanbul ignore next */
  throw new Error("Value " + JSON.stringify(val) + " is not a valid document");
}

/**
 * @param {Doc[]} parts
 * @returns Doc
 */
function concat(parts) {
  for (const part of parts) {
    assertDoc(part);
  }

  // We cannot do this until we change `printJSXElement` to not
  // access the internals of a document directly.
  // if(parts.length === 1) {
  //   // If it's a single document, no need to concat it.
  //   return parts[0];
  // }
  return {type: "concat", parts};
}

/**
 * @param {Doc} contents
 * @returns Doc
 */
function indent(contents) {
  assertDoc(contents);
  return {type: "indent", contents};
}

/**
 * @param {number | string} widthOrString
 * @param {Doc} contents
 * @returns Doc
 */
function align(widthOrString, contents) {
  assertDoc(contents);

  return {type: "align", contents, n: widthOrString};
}

/**
 * @param {Doc} contents
 * @param {object} [opts] - TBD ???
 * @returns Doc
 */
function group(contents, opts = {}) {
  assertDoc(contents);

  return {
    type: "group",
    id: opts.id,
    contents,
    break: Boolean(opts.shouldBreak),
    expandedStates: opts.expandedStates,
  };
}

/**
 * @param {Doc} contents
 * @returns Doc
 */
function dedentToRoot(contents) {
  return align(Number.NEGATIVE_INFINITY, contents);
}

/**
 * @param {Doc} contents
 * @returns Doc
 */
function markAsRoot(contents) {
  // @ts-expect-error - TBD ???:
  return align({type: "root"}, contents);
}

/**
 * @param {Doc} contents
 * @returns Doc
 */
function dedent(contents) {
  return align(-1, contents);
}

/**
 * @param {Doc[]} states
 * @param {object} [opts] - TBD ???
 * @returns Doc
 */
function conditionalGroup(states, opts) {
  return group(states[0], {...opts, expandedStates: states});
}

/**
 * @param {Doc[]} parts
 * @returns Doc
 */
function fill(parts) {
  for (const part of parts) {
    assertDoc(part);
  }

  return {type: "fill", parts};
}

/**
 * @param {Doc} [breakContents]
 * @param {Doc} [flatContents]
 * @param {object} [opts] - TBD ???
 * @returns Doc
 */
function ifBreak(breakContents, flatContents, opts = {}) {
  if (breakContents) {
    assertDoc(breakContents);
  }
  if (flatContents) {
    assertDoc(flatContents);
  }

  return {
    type: "if-break",
    breakContents,
    flatContents,
    groupId: opts.groupId,
  };
}

/**
 * Optimized version of `ifBreak(indent(doc), doc, { groupId: ... })`
 * @param {Doc} contents
 * @param {{ groupId: symbol, negate?: boolean }} opts
 * @returns Doc
 */
function indentIfBreak(contents, opts) {
  return {
    type: "indent-if-break",
    contents,
    groupId: opts.groupId,
    negate: opts.negate,
  };
}

/**
 * @param {Doc} contents
 * @returns Doc
 */
function lineSuffix(contents) {
  assertDoc(contents);
  return {type: "line-suffix", contents};
}

const lineSuffixBoundary = {type: "line-suffix-boundary"};
const breakParent = {type: "break-parent"};
const trim$1 = {type: "trim"};

const hardlineWithoutBreakParent = {type: "line", hard: true};
const literallineWithoutBreakParent = {
  type: "line",
  hard: true,
  literal: true,
};

const line = {type: "line"};
const softline = {type: "line", soft: true};
// eslint-disable-next-line prettier-internal-rules/no-doc-builder-concat
const hardline = concat([hardlineWithoutBreakParent, breakParent]);
// eslint-disable-next-line prettier-internal-rules/no-doc-builder-concat
const literalline = concat([literallineWithoutBreakParent, breakParent]);

const cursor = {type: "cursor", placeholder: Symbol("cursor")};

/**
 * @param {Doc} sep
 * @param {Doc[]} arr
 * @returns Doc
 */
function join(sep, arr) {
  const res = [];

  for (let i = 0; i < arr.length; i++) {
    if (i !== 0) {
      res.push(sep);
    }

    res.push(arr[i]);
  }

  // eslint-disable-next-line prettier-internal-rules/no-doc-builder-concat
  return concat(res);
}

/**
 * @param {Doc} doc
 * @param {number} size
 * @param {number} tabWidth
 */
function addAlignmentToDoc(doc, size, tabWidth) {
  let aligned = doc;
  if (size > 0) {
    // Use indent to add tabs for all the levels of tabs we need
    for (let i = 0; i < Math.floor(size / tabWidth); ++i) {
      aligned = indent(aligned);
    }
    // Use align for all the spaces that are needed
    aligned = align(size % tabWidth, aligned);
    // size is absolute from 0 and not relative to the current
    // indentation, so we use -Infinity to reset the indentation to 0
    aligned = align(Number.NEGATIVE_INFINITY, aligned);
  }
  return aligned;
}

function label(label, contents) {
  return {type: "label", label, contents};
}

var docBuilders = /*#__PURE__*/Object.freeze({
  __proto__: null,
  addAlignmentToDoc: addAlignmentToDoc,
  align: align,
  assertDoc: assertDoc,
  breakParent: breakParent,
  concat: concat,
  conditionalGroup: conditionalGroup,
  cursor: cursor,
  dedent: dedent,
  dedentToRoot: dedentToRoot,
  fill: fill,
  group: group,
  hardline: hardline,
  hardlineWithoutBreakParent: hardlineWithoutBreakParent,
  ifBreak: ifBreak,
  indent: indent,
  indentIfBreak: indentIfBreak,
  join: join,
  label: label,
  line: line,
  lineSuffix: lineSuffix,
  lineSuffixBoundary: lineSuffixBoundary,
  literalline: literalline,
  literallineWithoutBreakParent: literallineWithoutBreakParent,
  markAsRoot: markAsRoot,
  softline: softline,
  trim: trim$1
});

const EMOJI_REGEXP = /\uD83C\uDFF4\uDB40\uDC67\uDB40\uDC62(?:\uDB40\uDC77\uDB40\uDC6C\uDB40\uDC73|\uDB40\uDC73\uDB40\uDC63\uDB40\uDC74|\uDB40\uDC65\uDB40\uDC6E\uDB40\uDC67)\uDB40\uDC7F|(?:\uD83E\uDDD1\uD83C\uDFFF\u200D\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D)?\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFF\u200D\uD83E\uDD1D\u200D(?:\uD83D[\uDC68\uDC69]))(?:\uD83C[\uDFFB-\uDFFE])|(?:\uD83E\uDDD1\uD83C\uDFFE\u200D\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D)?\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFE\u200D\uD83E\uDD1D\u200D(?:\uD83D[\uDC68\uDC69]))(?:\uD83C[\uDFFB-\uDFFD\uDFFF])|(?:\uD83E\uDDD1\uD83C\uDFFD\u200D\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D)?\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFD\u200D\uD83E\uDD1D\u200D(?:\uD83D[\uDC68\uDC69]))(?:\uD83C[\uDFFB\uDFFC\uDFFE\uDFFF])|(?:\uD83E\uDDD1\uD83C\uDFFC\u200D\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D)?\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFC\u200D\uD83E\uDD1D\u200D(?:\uD83D[\uDC68\uDC69]))(?:\uD83C[\uDFFB\uDFFD-\uDFFF])|(?:\uD83E\uDDD1\uD83C\uDFFB\u200D\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D)?\uD83E\uDDD1|\uD83D\uDC69\uD83C\uDFFB\u200D\uD83E\uDD1D\u200D(?:\uD83D[\uDC68\uDC69]))(?:\uD83C[\uDFFC-\uDFFF])|\uD83D\uDC68(?:\uD83C\uDFFB(?:\u200D(?:\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFF])|\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFF]))|\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFC-\uDFFF])|[\u2695\u2696\u2708]\uFE0F|\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD]))?|(?:\uD83C[\uDFFC-\uDFFF])\u200D\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFF])|\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFF]))|\u200D(?:\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D)?\uD83D\uDC68|(?:\uD83D[\uDC68\uDC69])\u200D(?:\uD83D\uDC66\u200D\uD83D\uDC66|\uD83D\uDC67\u200D(?:\uD83D[\uDC66\uDC67]))|\uD83D\uDC66\u200D\uD83D\uDC66|\uD83D\uDC67\u200D(?:\uD83D[\uDC66\uDC67])|\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFF\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFE])|\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFE\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB-\uDFFD\uDFFF])|\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFD\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB\uDFFC\uDFFE\uDFFF])|\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFC\u200D(?:\uD83E\uDD1D\u200D\uD83D\uDC68(?:\uD83C[\uDFFB\uDFFD-\uDFFF])|\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|(?:\uD83C\uDFFF\u200D[\u2695\u2696\u2708]|\uD83C\uDFFE\u200D[\u2695\u2696\u2708]|\uD83C\uDFFD\u200D[\u2695\u2696\u2708]|\uD83C\uDFFC\u200D[\u2695\u2696\u2708]|\u200D[\u2695\u2696\u2708])\uFE0F|\u200D(?:(?:\uD83D[\uDC68\uDC69])\u200D(?:\uD83D[\uDC66\uDC67])|\uD83D[\uDC66\uDC67])|\uD83C\uDFFF|\uD83C\uDFFE|\uD83C\uDFFD|\uD83C\uDFFC)?|(?:\uD83D\uDC69(?:\uD83C\uDFFB\u200D\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D(?:\uD83D[\uDC68\uDC69])|\uD83D[\uDC68\uDC69])|(?:\uD83C[\uDFFC-\uDFFF])\u200D\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D(?:\uD83D[\uDC68\uDC69])|\uD83D[\uDC68\uDC69]))|\uD83E\uDDD1(?:\uD83C[\uDFFB-\uDFFF])\u200D\uD83E\uDD1D\u200D\uD83E\uDDD1)(?:\uD83C[\uDFFB-\uDFFF])|\uD83D\uDC69\u200D\uD83D\uDC69\u200D(?:\uD83D\uDC66\u200D\uD83D\uDC66|\uD83D\uDC67\u200D(?:\uD83D[\uDC66\uDC67]))|\uD83D\uDC69(?:\u200D(?:\u2764\uFE0F\u200D(?:\uD83D\uDC8B\u200D(?:\uD83D[\uDC68\uDC69])|\uD83D[\uDC68\uDC69])|\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFF\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFE\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFD\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFC\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFB\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD]))|\uD83E\uDDD1(?:\u200D(?:\uD83E\uDD1D\u200D\uD83E\uDDD1|\uD83C[\uDF3E\uDF73\uDF7C\uDF84\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFF\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF84\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFE\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF84\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFD\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF84\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFC\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF84\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD])|\uD83C\uDFFB\u200D(?:\uD83C[\uDF3E\uDF73\uDF7C\uDF84\uDF93\uDFA4\uDFA8\uDFEB\uDFED]|\uD83D[\uDCBB\uDCBC\uDD27\uDD2C\uDE80\uDE92]|\uD83E[\uDDAF-\uDDB3\uDDBC\uDDBD]))|\uD83D\uDC69\u200D\uD83D\uDC66\u200D\uD83D\uDC66|\uD83D\uDC69\u200D\uD83D\uDC69\u200D(?:\uD83D[\uDC66\uDC67])|\uD83D\uDC69\u200D\uD83D\uDC67\u200D(?:\uD83D[\uDC66\uDC67])|(?:\uD83D\uDC41\uFE0F\u200D\uD83D\uDDE8|\uD83E\uDDD1(?:\uD83C\uDFFF\u200D[\u2695\u2696\u2708]|\uD83C\uDFFE\u200D[\u2695\u2696\u2708]|\uD83C\uDFFD\u200D[\u2695\u2696\u2708]|\uD83C\uDFFC\u200D[\u2695\u2696\u2708]|\uD83C\uDFFB\u200D[\u2695\u2696\u2708]|\u200D[\u2695\u2696\u2708])|\uD83D\uDC69(?:\uD83C\uDFFF\u200D[\u2695\u2696\u2708]|\uD83C\uDFFE\u200D[\u2695\u2696\u2708]|\uD83C\uDFFD\u200D[\u2695\u2696\u2708]|\uD83C\uDFFC\u200D[\u2695\u2696\u2708]|\uD83C\uDFFB\u200D[\u2695\u2696\u2708]|\u200D[\u2695\u2696\u2708])|\uD83D\uDE36\u200D\uD83C\uDF2B|\uD83C\uDFF3\uFE0F\u200D\u26A7|\uD83D\uDC3B\u200D\u2744|(?:(?:\uD83C[\uDFC3\uDFC4\uDFCA]|\uD83D[\uDC6E\uDC70\uDC71\uDC73\uDC77\uDC81\uDC82\uDC86\uDC87\uDE45-\uDE47\uDE4B\uDE4D\uDE4E\uDEA3\uDEB4-\uDEB6]|\uD83E[\uDD26\uDD35\uDD37-\uDD39\uDD3D\uDD3E\uDDB8\uDDB9\uDDCD-\uDDCF\uDDD4\uDDD6-\uDDDD])(?:\uD83C[\uDFFB-\uDFFF])|\uD83D\uDC6F|\uD83E[\uDD3C\uDDDE\uDDDF])\u200D[\u2640\u2642]|(?:\u26F9|\uD83C[\uDFCB\uDFCC]|\uD83D\uDD75)(?:\uFE0F|\uD83C[\uDFFB-\uDFFF])\u200D[\u2640\u2642]|\uD83C\uDFF4\u200D\u2620|(?:\uD83C[\uDFC3\uDFC4\uDFCA]|\uD83D[\uDC6E\uDC70\uDC71\uDC73\uDC77\uDC81\uDC82\uDC86\uDC87\uDE45-\uDE47\uDE4B\uDE4D\uDE4E\uDEA3\uDEB4-\uDEB6]|\uD83E[\uDD26\uDD35\uDD37-\uDD39\uDD3D\uDD3E\uDDB8\uDDB9\uDDCD-\uDDCF\uDDD4\uDDD6-\uDDDD])\u200D[\u2640\u2642]|[\xA9\xAE\u203C\u2049\u2122\u2139\u2194-\u2199\u21A9\u21AA\u2328\u23CF\u23ED-\u23EF\u23F1\u23F2\u23F8-\u23FA\u24C2\u25AA\u25AB\u25B6\u25C0\u25FB\u25FC\u2600-\u2604\u260E\u2611\u2618\u2620\u2622\u2623\u2626\u262A\u262E\u262F\u2638-\u263A\u2640\u2642\u265F\u2660\u2663\u2665\u2666\u2668\u267B\u267E\u2692\u2694-\u2697\u2699\u269B\u269C\u26A0\u26A7\u26B0\u26B1\u26C8\u26CF\u26D1\u26D3\u26E9\u26F0\u26F1\u26F4\u26F7\u26F8\u2702\u2708\u2709\u270F\u2712\u2714\u2716\u271D\u2721\u2733\u2734\u2744\u2747\u2763\u27A1\u2934\u2935\u2B05-\u2B07\u3030\u303D\u3297\u3299]|\uD83C[\uDD70\uDD71\uDD7E\uDD7F\uDE02\uDE37\uDF21\uDF24-\uDF2C\uDF36\uDF7D\uDF96\uDF97\uDF99-\uDF9B\uDF9E\uDF9F\uDFCD\uDFCE\uDFD4-\uDFDF\uDFF5\uDFF7]|\uD83D[\uDC3F\uDCFD\uDD49\uDD4A\uDD6F\uDD70\uDD73\uDD76-\uDD79\uDD87\uDD8A-\uDD8D\uDDA5\uDDA8\uDDB1\uDDB2\uDDBC\uDDC2-\uDDC4\uDDD1-\uDDD3\uDDDC-\uDDDE\uDDE1\uDDE3\uDDE8\uDDEF\uDDF3\uDDFA\uDECB\uDECD-\uDECF\uDEE0-\uDEE5\uDEE9\uDEF0\uDEF3])\uFE0F|\uD83C\uDFF3\uFE0F\u200D\uD83C\uDF08|\uD83D\uDC69\u200D\uD83D\uDC67|\uD83D\uDC69\u200D\uD83D\uDC66|\uD83D\uDE35\u200D\uD83D\uDCAB|\uD83D\uDE2E\u200D\uD83D\uDCA8|\uD83D\uDC15\u200D\uD83E\uDDBA|\uD83E\uDDD1(?:\uD83C\uDFFF|\uD83C\uDFFE|\uD83C\uDFFD|\uD83C\uDFFC|\uD83C\uDFFB)?|\uD83D\uDC69(?:\uD83C\uDFFF|\uD83C\uDFFE|\uD83C\uDFFD|\uD83C\uDFFC|\uD83C\uDFFB)?|\uD83C\uDDFD\uD83C\uDDF0|\uD83C\uDDF6\uD83C\uDDE6|\uD83C\uDDF4\uD83C\uDDF2|\uD83D\uDC08\u200D\u2B1B|\u2764\uFE0F\u200D(?:\uD83D\uDD25|\uD83E\uDE79)|\uD83D\uDC41\uFE0F|\uD83C\uDFF3\uFE0F|\uD83C\uDDFF(?:\uD83C[\uDDE6\uDDF2\uDDFC])|\uD83C\uDDFE(?:\uD83C[\uDDEA\uDDF9])|\uD83C\uDDFC(?:\uD83C[\uDDEB\uDDF8])|\uD83C\uDDFB(?:\uD83C[\uDDE6\uDDE8\uDDEA\uDDEC\uDDEE\uDDF3\uDDFA])|\uD83C\uDDFA(?:\uD83C[\uDDE6\uDDEC\uDDF2\uDDF3\uDDF8\uDDFE\uDDFF])|\uD83C\uDDF9(?:\uD83C[\uDDE6\uDDE8\uDDE9\uDDEB-\uDDED\uDDEF-\uDDF4\uDDF7\uDDF9\uDDFB\uDDFC\uDDFF])|\uD83C\uDDF8(?:\uD83C[\uDDE6-\uDDEA\uDDEC-\uDDF4\uDDF7-\uDDF9\uDDFB\uDDFD-\uDDFF])|\uD83C\uDDF7(?:\uD83C[\uDDEA\uDDF4\uDDF8\uDDFA\uDDFC])|\uD83C\uDDF5(?:\uD83C[\uDDE6\uDDEA-\uDDED\uDDF0-\uDDF3\uDDF7-\uDDF9\uDDFC\uDDFE])|\uD83C\uDDF3(?:\uD83C[\uDDE6\uDDE8\uDDEA-\uDDEC\uDDEE\uDDF1\uDDF4\uDDF5\uDDF7\uDDFA\uDDFF])|\uD83C\uDDF2(?:\uD83C[\uDDE6\uDDE8-\uDDED\uDDF0-\uDDFF])|\uD83C\uDDF1(?:\uD83C[\uDDE6-\uDDE8\uDDEE\uDDF0\uDDF7-\uDDFB\uDDFE])|\uD83C\uDDF0(?:\uD83C[\uDDEA\uDDEC-\uDDEE\uDDF2\uDDF3\uDDF5\uDDF7\uDDFC\uDDFE\uDDFF])|\uD83C\uDDEF(?:\uD83C[\uDDEA\uDDF2\uDDF4\uDDF5])|\uD83C\uDDEE(?:\uD83C[\uDDE8-\uDDEA\uDDF1-\uDDF4\uDDF6-\uDDF9])|\uD83C\uDDED(?:\uD83C[\uDDF0\uDDF2\uDDF3\uDDF7\uDDF9\uDDFA])|\uD83C\uDDEC(?:\uD83C[\uDDE6\uDDE7\uDDE9-\uDDEE\uDDF1-\uDDF3\uDDF5-\uDDFA\uDDFC\uDDFE])|\uD83C\uDDEB(?:\uD83C[\uDDEE-\uDDF0\uDDF2\uDDF4\uDDF7])|\uD83C\uDDEA(?:\uD83C[\uDDE6\uDDE8\uDDEA\uDDEC\uDDED\uDDF7-\uDDFA])|\uD83C\uDDE9(?:\uD83C[\uDDEA\uDDEC\uDDEF\uDDF0\uDDF2\uDDF4\uDDFF])|\uD83C\uDDE8(?:\uD83C[\uDDE6\uDDE8\uDDE9\uDDEB-\uDDEE\uDDF0-\uDDF5\uDDF7\uDDFA-\uDDFF])|\uD83C\uDDE7(?:\uD83C[\uDDE6\uDDE7\uDDE9-\uDDEF\uDDF1-\uDDF4\uDDF6-\uDDF9\uDDFB\uDDFC\uDDFE\uDDFF])|\uD83C\uDDE6(?:\uD83C[\uDDE8-\uDDEC\uDDEE\uDDF1\uDDF2\uDDF4\uDDF6-\uDDFA\uDDFC\uDDFD\uDDFF])|[#\*0-9]\uFE0F\u20E3|\u2764\uFE0F|(?:\uD83C[\uDFC3\uDFC4\uDFCA]|\uD83D[\uDC6E\uDC70\uDC71\uDC73\uDC77\uDC81\uDC82\uDC86\uDC87\uDE45-\uDE47\uDE4B\uDE4D\uDE4E\uDEA3\uDEB4-\uDEB6]|\uD83E[\uDD26\uDD35\uDD37-\uDD39\uDD3D\uDD3E\uDDB8\uDDB9\uDDCD-\uDDCF\uDDD4\uDDD6-\uDDDD])(?:\uD83C[\uDFFB-\uDFFF])|(?:\u26F9|\uD83C[\uDFCB\uDFCC]|\uD83D\uDD75)(?:\uFE0F|\uD83C[\uDFFB-\uDFFF])|\uD83C\uDFF4|(?:[\u270A\u270B]|\uD83C[\uDF85\uDFC2\uDFC7]|\uD83D[\uDC42\uDC43\uDC46-\uDC50\uDC66\uDC67\uDC6B-\uDC6D\uDC72\uDC74-\uDC76\uDC78\uDC7C\uDC83\uDC85\uDC8F\uDC91\uDCAA\uDD7A\uDD95\uDD96\uDE4C\uDE4F\uDEC0\uDECC]|\uD83E[\uDD0C\uDD0F\uDD18-\uDD1C\uDD1E\uDD1F\uDD30-\uDD34\uDD36\uDD77\uDDB5\uDDB6\uDDBB\uDDD2\uDDD3\uDDD5])(?:\uD83C[\uDFFB-\uDFFF])|(?:[\u261D\u270C\u270D]|\uD83D[\uDD74\uDD90])(?:\uFE0F|\uD83C[\uDFFB-\uDFFF])|[\u270A\u270B]|\uD83C[\uDF85\uDFC2\uDFC7]|\uD83D[\uDC08\uDC15\uDC3B\uDC42\uDC43\uDC46-\uDC50\uDC66\uDC67\uDC6B-\uDC6D\uDC72\uDC74-\uDC76\uDC78\uDC7C\uDC83\uDC85\uDC8F\uDC91\uDCAA\uDD7A\uDD95\uDD96\uDE2E\uDE35\uDE36\uDE4C\uDE4F\uDEC0\uDECC]|\uD83E[\uDD0C\uDD0F\uDD18-\uDD1C\uDD1E\uDD1F\uDD30-\uDD34\uDD36\uDD77\uDDB5\uDDB6\uDDBB\uDDD2\uDDD3\uDDD5]|\uD83C[\uDFC3\uDFC4\uDFCA]|\uD83D[\uDC6E\uDC70\uDC71\uDC73\uDC77\uDC81\uDC82\uDC86\uDC87\uDE45-\uDE47\uDE4B\uDE4D\uDE4E\uDEA3\uDEB4-\uDEB6]|\uD83E[\uDD26\uDD35\uDD37-\uDD39\uDD3D\uDD3E\uDDB8\uDDB9\uDDCD-\uDDCF\uDDD4\uDDD6-\uDDDD]|\uD83D\uDC6F|\uD83E[\uDD3C\uDDDE\uDDDF]|[\u231A\u231B\u23E9-\u23EC\u23F0\u23F3\u25FD\u25FE\u2614\u2615\u2648-\u2653\u267F\u2693\u26A1\u26AA\u26AB\u26BD\u26BE\u26C4\u26C5\u26CE\u26D4\u26EA\u26F2\u26F3\u26F5\u26FA\u26FD\u2705\u2728\u274C\u274E\u2753-\u2755\u2757\u2795-\u2797\u27B0\u27BF\u2B1B\u2B1C\u2B50\u2B55]|\uD83C[\uDC04\uDCCF\uDD8E\uDD91-\uDD9A\uDE01\uDE1A\uDE2F\uDE32-\uDE36\uDE38-\uDE3A\uDE50\uDE51\uDF00-\uDF20\uDF2D-\uDF35\uDF37-\uDF7C\uDF7E-\uDF84\uDF86-\uDF93\uDFA0-\uDFC1\uDFC5\uDFC6\uDFC8\uDFC9\uDFCF-\uDFD3\uDFE0-\uDFF0\uDFF8-\uDFFF]|\uD83D[\uDC00-\uDC07\uDC09-\uDC14\uDC16-\uDC3A\uDC3C-\uDC3E\uDC40\uDC44\uDC45\uDC51-\uDC65\uDC6A\uDC79-\uDC7B\uDC7D-\uDC80\uDC84\uDC88-\uDC8E\uDC90\uDC92-\uDCA9\uDCAB-\uDCFC\uDCFF-\uDD3D\uDD4B-\uDD4E\uDD50-\uDD67\uDDA4\uDDFB-\uDE2D\uDE2F-\uDE34\uDE37-\uDE44\uDE48-\uDE4A\uDE80-\uDEA2\uDEA4-\uDEB3\uDEB7-\uDEBF\uDEC1-\uDEC5\uDED0-\uDED2\uDED5-\uDED7\uDEEB\uDEEC\uDEF4-\uDEFC\uDFE0-\uDFEB]|\uD83E[\uDD0D\uDD0E\uDD10-\uDD17\uDD1D\uDD20-\uDD25\uDD27-\uDD2F\uDD3A\uDD3F-\uDD45\uDD47-\uDD76\uDD78\uDD7A-\uDDB4\uDDB7\uDDBA\uDDBC-\uDDCB\uDDD0\uDDE0-\uDDFF\uDE70-\uDE74\uDE78-\uDE7A\uDE80-\uDE86\uDE90-\uDEA8\uDEB0-\uDEB6\uDEC0-\uDEC2\uDED0-\uDED6]|(?:[\u231A\u231B\u23E9-\u23EC\u23F0\u23F3\u25FD\u25FE\u2614\u2615\u2648-\u2653\u267F\u2693\u26A1\u26AA\u26AB\u26BD\u26BE\u26C4\u26C5\u26CE\u26D4\u26EA\u26F2\u26F3\u26F5\u26FA\u26FD\u2705\u270A\u270B\u2728\u274C\u274E\u2753-\u2755\u2757\u2795-\u2797\u27B0\u27BF\u2B1B\u2B1C\u2B50\u2B55]|\uD83C[\uDC04\uDCCF\uDD8E\uDD91-\uDD9A\uDDE6-\uDDFF\uDE01\uDE1A\uDE2F\uDE32-\uDE36\uDE38-\uDE3A\uDE50\uDE51\uDF00-\uDF20\uDF2D-\uDF35\uDF37-\uDF7C\uDF7E-\uDF93\uDFA0-\uDFCA\uDFCF-\uDFD3\uDFE0-\uDFF0\uDFF4\uDFF8-\uDFFF]|\uD83D[\uDC00-\uDC3E\uDC40\uDC42-\uDCFC\uDCFF-\uDD3D\uDD4B-\uDD4E\uDD50-\uDD67\uDD7A\uDD95\uDD96\uDDA4\uDDFB-\uDE4F\uDE80-\uDEC5\uDECC\uDED0-\uDED2\uDED5-\uDED7\uDEEB\uDEEC\uDEF4-\uDEFC\uDFE0-\uDFEB]|\uD83E[\uDD0C-\uDD3A\uDD3C-\uDD45\uDD47-\uDD78\uDD7A-\uDDCB\uDDCD-\uDDFF\uDE70-\uDE74\uDE78-\uDE7A\uDE80-\uDE86\uDE90-\uDEA8\uDEB0-\uDEB6\uDEC0-\uDEC2\uDED0-\uDED6])|(?:[#\*0-9\xA9\xAE\u203C\u2049\u2122\u2139\u2194-\u2199\u21A9\u21AA\u231A\u231B\u2328\u23CF\u23E9-\u23F3\u23F8-\u23FA\u24C2\u25AA\u25AB\u25B6\u25C0\u25FB-\u25FE\u2600-\u2604\u260E\u2611\u2614\u2615\u2618\u261D\u2620\u2622\u2623\u2626\u262A\u262E\u262F\u2638-\u263A\u2640\u2642\u2648-\u2653\u265F\u2660\u2663\u2665\u2666\u2668\u267B\u267E\u267F\u2692-\u2697\u2699\u269B\u269C\u26A0\u26A1\u26A7\u26AA\u26AB\u26B0\u26B1\u26BD\u26BE\u26C4\u26C5\u26C8\u26CE\u26CF\u26D1\u26D3\u26D4\u26E9\u26EA\u26F0-\u26F5\u26F7-\u26FA\u26FD\u2702\u2705\u2708-\u270D\u270F\u2712\u2714\u2716\u271D\u2721\u2728\u2733\u2734\u2744\u2747\u274C\u274E\u2753-\u2755\u2757\u2763\u2764\u2795-\u2797\u27A1\u27B0\u27BF\u2934\u2935\u2B05-\u2B07\u2B1B\u2B1C\u2B50\u2B55\u3030\u303D\u3297\u3299]|\uD83C[\uDC04\uDCCF\uDD70\uDD71\uDD7E\uDD7F\uDD8E\uDD91-\uDD9A\uDDE6-\uDDFF\uDE01\uDE02\uDE1A\uDE2F\uDE32-\uDE3A\uDE50\uDE51\uDF00-\uDF21\uDF24-\uDF93\uDF96\uDF97\uDF99-\uDF9B\uDF9E-\uDFF0\uDFF3-\uDFF5\uDFF7-\uDFFF]|\uD83D[\uDC00-\uDCFD\uDCFF-\uDD3D\uDD49-\uDD4E\uDD50-\uDD67\uDD6F\uDD70\uDD73-\uDD7A\uDD87\uDD8A-\uDD8D\uDD90\uDD95\uDD96\uDDA4\uDDA5\uDDA8\uDDB1\uDDB2\uDDBC\uDDC2-\uDDC4\uDDD1-\uDDD3\uDDDC-\uDDDE\uDDE1\uDDE3\uDDE8\uDDEF\uDDF3\uDDFA-\uDE4F\uDE80-\uDEC5\uDECB-\uDED2\uDED5-\uDED7\uDEE0-\uDEE5\uDEE9\uDEEB\uDEEC\uDEF0\uDEF3-\uDEFC\uDFE0-\uDFEB]|\uD83E[\uDD0C-\uDD3A\uDD3C-\uDD45\uDD47-\uDD78\uDD7A-\uDDCB\uDDCD-\uDDFF\uDE70-\uDE74\uDE78-\uDE7A\uDE80-\uDE86\uDE90-\uDEA8\uDEB0-\uDEB6\uDEC0-\uDEC2\uDED0-\uDED6])\uFE0F|(?:[\u261D\u26F9\u270A-\u270D]|\uD83C[\uDF85\uDFC2-\uDFC4\uDFC7\uDFCA-\uDFCC]|\uD83D[\uDC42\uDC43\uDC46-\uDC50\uDC66-\uDC78\uDC7C\uDC81-\uDC83\uDC85-\uDC87\uDC8F\uDC91\uDCAA\uDD74\uDD75\uDD7A\uDD90\uDD95\uDD96\uDE45-\uDE47\uDE4B-\uDE4F\uDEA3\uDEB4-\uDEB6\uDEC0\uDECC]|\uD83E[\uDD0C\uDD0F\uDD18-\uDD1F\uDD26\uDD30-\uDD39\uDD3C-\uDD3E\uDD77\uDDB5\uDDB6\uDDB8\uDDB9\uDDBB\uDDCD-\uDDCF\uDDD1-\uDDDD])/g;

// node_modules/strip-ansi/node_modules/ansi-regex/index.js
function ansiRegex({onlyFirst = false} = {}) {
  const pattern = [
    "[\\u001B\\u009B][[\\]()#;?]*(?:(?:(?:(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]+)*|[a-zA-Z\\d]+(?:;[-a-zA-Z\\d\\/#&.:=?%@~_]*)*)?\\u0007)",
    "(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-ntqry=><~]))"
  ].join("|");
  return new RegExp(pattern, onlyFirst ? void 0 : "g");
}

// node_modules/strip-ansi/index.js
function stripAnsi(string) {
  if (typeof string !== "string") {
    throw new TypeError(`Expected a \`string\`, got \`${typeof string}\``);
  }
  return string.replace(ansiRegex(), "");
}

// node_modules/is-fullwidth-code-point/index.js
function isFullwidthCodePoint(codePoint) {
  if (!Number.isInteger(codePoint)) {
    return false;
  }
  return codePoint >= 4352 && (codePoint <= 4447 || codePoint === 9001 || codePoint === 9002 || 11904 <= codePoint && codePoint <= 12871 && codePoint !== 12351 || 12880 <= codePoint && codePoint <= 19903 || 19968 <= codePoint && codePoint <= 42182 || 43360 <= codePoint && codePoint <= 43388 || 44032 <= codePoint && codePoint <= 55203 || 63744 <= codePoint && codePoint <= 64255 || 65040 <= codePoint && codePoint <= 65049 || 65072 <= codePoint && codePoint <= 65131 || 65281 <= codePoint && codePoint <= 65376 || 65504 <= codePoint && codePoint <= 65510 || 110592 <= codePoint && codePoint <= 110593 || 127488 <= codePoint && codePoint <= 127569 || 131072 <= codePoint && codePoint <= 262141);
}

// node_modules/string-width/index.js
function stringWidth(string) {
  if (typeof string !== "string" || string.length === 0) {
    return 0;
  }
  string = stripAnsi(string);
  if (string.length === 0) {
    return 0;
  }
  string = string.replace(EMOJI_REGEXP, "  ");
  let width = 0;
  for (let index = 0; index < string.length; index++) {
    const codePoint = string.codePointAt(index);
    if (codePoint <= 31 || codePoint >= 127 && codePoint <= 159) {
      continue;
    }
    if (codePoint >= 768 && codePoint <= 879) {
      continue;
    }
    if (codePoint > 65535) {
      index++;
    }
    width += isFullwidthCodePoint(codePoint) ? 2 : 1;
  }
  return width;
}

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
    this.doc = doc;
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
              if (typeof value === 'string') value = `"${value}"`;
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
      this.pushImportTypes(typeName);
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
      this.pushImportTypes(typeName);
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
    this.importedTypes = [...new Set([...this.importedTypes, typeName])];
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
    console.log('unknown type', 'schema.type=' + t.type, 'ext.type=' + t.ext?.type);
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
            const p = properties[k];
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

function convertEndOfLineToChars(value) {
  switch (value) {
    case "cr":
      return "\r";
    case "crlf":
      return "\r\n";
    default:
      return "\n";
  }
}

const getLast = (arr) => arr[arr.length - 1];

const notAsciiRegex = /[^\x20-\x7F]/;

/**
 * @param {string} text
 * @returns {number}
 */
function getStringWidth(text) {
  if (!text) {
    return 0;
  }

  // shortcut to avoid needless string `RegExp`s, replacements, and allocations within `string-width`
  if (!notAsciiRegex.test(text)) {
    return text.length;
  }

  return stringWidth(text);
}

const isConcat = (doc) => Array.isArray(doc) || (doc && doc.type === "concat");
const getDocParts = (doc) => {
  if (Array.isArray(doc)) {
    return doc;
  }

  /* istanbul ignore next */
  if (doc.type !== "concat" && doc.type !== "fill") {
    throw new Error("Expect doc type to be `concat` or `fill`.");
  }

  return doc.parts;
};

/** @typedef {typeof MODE_BREAK | typeof MODE_FLAT} Mode */
/** @typedef {{ ind: any, doc: any, mode: Mode }} Command */

/** @type {Record<symbol, Mode>} */
let groupModeMap;

// prettier-ignore
const MODE_BREAK = /** @type {const} */ (1);
// prettier-ignore
const MODE_FLAT = /** @type {const} */ (2);

function rootIndent() {
  return {value: "", length: 0, queue: []};
}

function makeIndent(ind, options) {
  return generateInd(ind, {type: "indent"}, options);
}

function makeAlign(indent, widthOrDoc, options) {
  if (widthOrDoc === Number.NEGATIVE_INFINITY) {
    return indent.root || rootIndent();
  }

  if (widthOrDoc < 0) {
    return generateInd(indent, {type: "dedent"}, options);
  }

  if (!widthOrDoc) {
    return indent;
  }

  if (widthOrDoc.type === "root") {
    return {...indent, root: indent};
  }

  const alignType =
    typeof widthOrDoc === "string" ? "stringAlign" : "numberAlign";

  return generateInd(indent, {type: alignType, n: widthOrDoc}, options);
}

function generateInd(ind, newPart, options) {
  const queue =
    newPart.type === "dedent"
      ? ind.queue.slice(0, -1)
      : [...ind.queue, newPart];

  let value = "";
  let length = 0;
  let lastTabs = 0;
  let lastSpaces = 0;

  for (const part of queue) {
    switch (part.type) {
      case "indent":
        flush();
        if (options.useTabs) {
          addTabs(1);
        } else {
          addSpaces(options.tabWidth);
        }
        break;
      case "stringAlign":
        flush();
        value += part.n;
        length += part.n.length;
        break;
      case "numberAlign":
        lastTabs += 1;
        lastSpaces += part.n;
        break;
      /* istanbul ignore next */
      default:
        throw new Error(`Unexpected type '${part.type}'`);
    }
  }

  flushSpaces();

  return {...ind, value, length, queue};

  function addTabs(count) {
    value += "\t".repeat(count);
    length += options.tabWidth * count;
  }

  function addSpaces(count) {
    value += " ".repeat(count);
    length += count;
  }

  function flush() {
    if (options.useTabs) {
      flushTabs();
    } else {
      flushSpaces();
    }
  }

  function flushTabs() {
    if (lastTabs > 0) {
      addTabs(lastTabs);
    }
    resetLast();
  }

  function flushSpaces() {
    if (lastSpaces > 0) {
      addSpaces(lastSpaces);
    }
    resetLast();
  }

  function resetLast() {
    lastTabs = 0;
    lastSpaces = 0;
  }
}

function trim(out) {
  if (out.length === 0) {
    return 0;
  }

  let trimCount = 0;

  // Trim whitespace at the end of line
  while (
    out.length > 0 &&
    typeof getLast(out) === "string" &&
    /^[\t ]*$/.test(getLast(out))
    ) {
    trimCount += out.pop().length;
  }

  if (out.length > 0 && typeof getLast(out) === "string") {
    const trimmed = getLast(out).replace(/[\t ]*$/, "");
    trimCount += getLast(out).length - trimmed.length;
    out[out.length - 1] = trimmed;
  }

  return trimCount;
}

/**
 * @param {Command} next
 * @param {Command[]} restCommands
 * @param {number} width
 * @param {boolean} hasLineSuffix
 * @param {boolean} [mustBeFlat]
 * @returns {boolean}
 */
function fits(next, restCommands, width, hasLineSuffix, mustBeFlat) {
  let restIdx = restCommands.length;
  /** @type {Array<Omit<Command, 'ind'>>} */
  const cmds = [next];
  // `out` is only used for width counting because `trim` requires to look
  // backwards for space characters.
  const out = [];
  while (width >= 0) {
    if (cmds.length === 0) {
      if (restIdx === 0) {
        return true;
      }
      cmds.push(restCommands[--restIdx]);
      continue;
    }

    const {mode, doc} = cmds.pop();

    if (typeof doc === "string") {
      out.push(doc);
      width -= getStringWidth(doc);
    } else if (isConcat(doc) || doc.type === "fill") {
      const parts = getDocParts(doc);
      for (let i = parts.length - 1; i >= 0; i--) {
        cmds.push({mode, doc: parts[i]});
      }
    } else {
      switch (doc.type) {
        case "indent":
        case "align":
        case "indent-if-break":
        case "label":
          cmds.push({mode, doc: doc.contents});
          break;

        case "trim":
          width += trim(out);
          break;

        case "group": {
          if (mustBeFlat && doc.break) {
            return false;
          }
          const groupMode = doc.break ? MODE_BREAK : mode;
          // The most expanded state takes up the least space on the current line.
          const contents =
            doc.expandedStates && groupMode === MODE_BREAK
              ? getLast(doc.expandedStates)
              : doc.contents;
          cmds.push({mode: groupMode, doc: contents});
          break;
        }

        case "if-break": {
          const groupMode = doc.groupId
            ? groupModeMap[doc.groupId] || MODE_FLAT
            : mode;
          const contents =
            groupMode === MODE_BREAK ? doc.breakContents : doc.flatContents;
          if (contents) {
            cmds.push({mode, doc: contents});
          }
          break;
        }

        case "line":
          if (mode === MODE_BREAK || doc.hard) {
            return true;
          }
          if (!doc.soft) {
            out.push(" ");
            width--;
          }
          break;

        case "line-suffix":
          hasLineSuffix = true;
          break;

        case "line-suffix-boundary":
          if (hasLineSuffix) {
            return false;
          }
          break;
      }
    }
  }
  return false;
}

function printDocToString(doc, options) {
  groupModeMap = {};

  const width = options.printWidth;
  const newLine = convertEndOfLineToChars(options.endOfLine);
  let pos = 0;
  // cmds is basically a stack. We've turned a recursive call into a
  // while loop which is much faster. The while loop below adds new
  // cmds to the array instead of recursively calling `print`.
  /** @type Command[] */
  const cmds = [{ind: rootIndent(), mode: MODE_BREAK, doc}];
  const out = [];
  let shouldRemeasure = false;
  /** @type Command[] */
  const lineSuffix = [];

  while (cmds.length > 0) {
    const {ind, mode, doc} = cmds.pop();

    if (typeof doc === "string") {
      const formatted = newLine !== "\n" ? doc.replace(/\n/g, newLine) : doc;
      out.push(formatted);
      pos += getStringWidth(formatted);
    } else if (isConcat(doc)) {
      const parts = getDocParts(doc);
      for (let i = parts.length - 1; i >= 0; i--) {
        cmds.push({ind, mode, doc: parts[i]});
      }
    } else {
      switch (doc.type) {
        case "cursor":
          out.push(cursor.placeholder);

          break;
        case "indent":
          cmds.push({ind: makeIndent(ind, options), mode, doc: doc.contents});

          break;
        case "align":
          cmds.push({
            ind: makeAlign(ind, doc.n, options),
            mode,
            doc: doc.contents,
          });

          break;
        case "trim":
          pos -= trim(out);

          break;
        case "group":
          switch (mode) {
            case MODE_FLAT:
              if (!shouldRemeasure) {
                cmds.push({
                  ind,
                  mode: doc.break ? MODE_BREAK : MODE_FLAT,
                  doc: doc.contents,
                });

                break;
              }
            // fallthrough

            case MODE_BREAK: {
              shouldRemeasure = false;

              const next = {ind, mode: MODE_FLAT, doc: doc.contents};
              const rem = width - pos;
              const hasLineSuffix = lineSuffix.length > 0;

              if (!doc.break && fits(next, cmds, rem, hasLineSuffix)) {
                cmds.push(next);
              } else {
                // Expanded states are a rare case where a document
                // can manually provide multiple representations of
                // itself. It provides an array of documents
                // going from the least expanded (most flattened)
                // representation first to the most expanded. If a
                // group has these, we need to manually go through
                // these states and find the first one that fits.
                if (doc.expandedStates) {
                  const mostExpanded = getLast(doc.expandedStates);

                  if (doc.break) {
                    cmds.push({ind, mode: MODE_BREAK, doc: mostExpanded});

                    break;
                  } else {
                    for (let i = 1; i < doc.expandedStates.length + 1; i++) {
                      if (i >= doc.expandedStates.length) {
                        cmds.push({ind, mode: MODE_BREAK, doc: mostExpanded});

                        break;
                      } else {
                        const state = doc.expandedStates[i];
                        const cmd = {ind, mode: MODE_FLAT, doc: state};

                        if (fits(cmd, cmds, rem, hasLineSuffix)) {
                          cmds.push(cmd);

                          break;
                        }
                      }
                    }
                  }
                } else {
                  cmds.push({ind, mode: MODE_BREAK, doc: doc.contents});
                }
              }

              break;
            }
          }

          if (doc.id) {
            groupModeMap[doc.id] = getLast(cmds).mode;
          }
          break;
        // Fills each line with as much code as possible before moving to a new
        // line with the same indentation.
        //
        // Expects doc.parts to be an array of alternating content and
        // whitespace. The whitespace contains the linebreaks.
        //
        // For example:
        //   ["I", line, "love", line, "monkeys"]
        // or
        //   [{ type: group, ... }, softline, { type: group, ... }]
        //
        // It uses this parts structure to handle three main layout cases:
        // * The first two content items fit on the same line without
        //   breaking
        //   -> output the first content item and the whitespace "flat".
        // * Only the first content item fits on the line without breaking
        //   -> output the first content item "flat" and the whitespace with
        //   "break".
        // * Neither content item fits on the line without breaking
        //   -> output the first content item and the whitespace with "break".
        case "fill": {
          const rem = width - pos;

          const {parts} = doc;
          if (parts.length === 0) {
            break;
          }

          const [content, whitespace] = parts;
          const contentFlatCmd = {ind, mode: MODE_FLAT, doc: content};
          const contentBreakCmd = {ind, mode: MODE_BREAK, doc: content};
          const contentFits = fits(
            contentFlatCmd,
            [],
            rem,
            lineSuffix.length > 0,
            true
          );

          if (parts.length === 1) {
            if (contentFits) {
              cmds.push(contentFlatCmd);
            } else {
              cmds.push(contentBreakCmd);
            }
            break;
          }

          const whitespaceFlatCmd = {ind, mode: MODE_FLAT, doc: whitespace};
          const whitespaceBreakCmd = {ind, mode: MODE_BREAK, doc: whitespace};

          if (parts.length === 2) {
            if (contentFits) {
              cmds.push(whitespaceFlatCmd, contentFlatCmd);
            } else {
              cmds.push(whitespaceBreakCmd, contentBreakCmd);
            }
            break;
          }

          // At this point we've handled the first pair (context, separator)
          // and will create a new fill doc for the rest of the content.
          // Ideally we wouldn't mutate the array here but copying all the
          // elements to a new array would make this algorithm quadratic,
          // which is unusable for large arrays (e.g. large texts in JSX).
          parts.splice(0, 2);
          const remainingCmd = {ind, mode, doc: fill(parts)};

          const secondContent = parts[0];

          const firstAndSecondContentFlatCmd = {
            ind,
            mode: MODE_FLAT,
            doc: [content, whitespace, secondContent],
          };
          const firstAndSecondContentFits = fits(
            firstAndSecondContentFlatCmd,
            [],
            rem,
            lineSuffix.length > 0,
            true
          );

          if (firstAndSecondContentFits) {
            cmds.push(remainingCmd, whitespaceFlatCmd, contentFlatCmd);
          } else if (contentFits) {
            cmds.push(remainingCmd, whitespaceBreakCmd, contentFlatCmd);
          } else {
            cmds.push(remainingCmd, whitespaceBreakCmd, contentBreakCmd);
          }
          break;
        }
        case "if-break":
        case "indent-if-break": {
          const groupMode = doc.groupId ? groupModeMap[doc.groupId] : mode;
          if (groupMode === MODE_BREAK) {
            const breakContents =
              doc.type === "if-break"
                ? doc.breakContents
                : doc.negate
                  ? doc.contents
                  : indent(doc.contents);
            if (breakContents) {
              cmds.push({ind, mode, doc: breakContents});
            }
          }
          if (groupMode === MODE_FLAT) {
            const flatContents =
              doc.type === "if-break"
                ? doc.flatContents
                : doc.negate
                  ? indent(doc.contents)
                  : doc.contents;
            if (flatContents) {
              cmds.push({ind, mode, doc: flatContents});
            }
          }

          break;
        }
        case "line-suffix":
          lineSuffix.push({ind, mode, doc: doc.contents});
          break;
        case "line-suffix-boundary":
          if (lineSuffix.length > 0) {
            cmds.push({ind, mode, doc: {type: "line", hard: true}});
          }
          break;
        case "line":
          switch (mode) {
            case MODE_FLAT:
              if (!doc.hard) {
                if (!doc.soft) {
                  out.push(" ");

                  pos += 1;
                }

                break;
              } else {
                // This line was forced into the output even if we
                // were in flattened mode, so we need to tell the next
                // group that no matter what, it needs to remeasure
                // because the previous measurement didn't accurately
                // capture the entire expression (this is necessary
                // for nested groups)
                shouldRemeasure = true;
              }
            // fallthrough

            case MODE_BREAK:
              if (lineSuffix.length > 0) {
                cmds.push({ind, mode, doc}, ...lineSuffix.reverse());
                lineSuffix.length = 0;
                break;
              }

              if (doc.literal) {
                if (ind.root) {
                  out.push(newLine, ind.root.value);
                  pos = ind.root.length;
                } else {
                  out.push(newLine);
                  pos = 0;
                }
              } else {
                pos -= trim(out);
                out.push(newLine + ind.value);
                pos = ind.length;
              }
              break;
          }
          break;
        case "label":
          cmds.push({ind, mode, doc: doc.contents});
          break;
      }
    }

    // Flush remaining line-suffix contents at the end of the document, in case
    // there is no new line after the line-suffix.
    if (cmds.length === 0 && lineSuffix.length > 0) {
      cmds.push(...lineSuffix.reverse());
      lineSuffix.length = 0;
    }
  }

  const cursorPlaceholderIndex = out.indexOf(cursor.placeholder);
  if (cursorPlaceholderIndex !== -1) {
    const otherCursorPlaceholderIndex = out.indexOf(
      cursor.placeholder,
      cursorPlaceholderIndex + 1
    );
    const beforeCursor = out.slice(0, cursorPlaceholderIndex).join("");
    const aroundCursor = out
      .slice(cursorPlaceholderIndex + 1, otherCursorPlaceholderIndex)
      .join("");
    const afterCursor = out.slice(otherCursorPlaceholderIndex + 1).join("");

    return {
      formatted: beforeCursor + aroundCursor + afterCursor,
      cursorNodeStart: beforeCursor.length,
      cursorNodeText: aroundCursor,
    };
  }

  return {formatted: out.join("")};
}

const TOKEN_SEPARATOR = 'separator';
const TOKEN_UPPER_CASE = 'upper_case';
const TOKEN_LOWER_CASE = 'lower_case';
const TOKEN_NUMBER = 'number';

const MATCH_RULES = [
  {type: TOKEN_SEPARATOR, pattern: /^[_\.\-]+/},
  {type: TOKEN_UPPER_CASE, pattern: /^[A-Z]+/},
  {type: TOKEN_LOWER_CASE, pattern: /^[a-z]+/},
  {type: TOKEN_NUMBER, pattern: /^\d+/},
];

/**
 * @typedef {object} Options
 * @property {boolean} pascalCase
 *
 * @param {string} input
 * @param {Options} options
 */
function camelCase(input, options) {
  let res = '';
  let cursor = 0;
  let wordStart = !!options?.pascalCase;

  while (cursor < input.length) {
    const subStr = input.substr(cursor);
    let matchedStr = null, matchedToken = null;
    for (const rule of MATCH_RULES) {
      const matched = subStr.match(rule.pattern);
      if (!matched) continue;
      matchedStr = matched[0];
      matchedToken = rule.type;
      break
    }

    if (!matchedStr) { // not matched
      res += subStr.charAt(0);
      cursor += 1;
      continue
    }

    cursor += matchedStr.length;
    switch (matchedToken) {
      case TOKEN_SEPARATOR:
        wordStart = true;
        break

      case TOKEN_UPPER_CASE:
        res += matchedStr;
        wordStart = false;
        break;

      case TOKEN_LOWER_CASE:
      case TOKEN_NUMBER:
        if (wordStart) {
          matchedStr = matchedStr[0].toUpperCase() + matchedStr.substr(1);
          wordStart = false;
        }
        res += matchedStr;
        break
    }
  }

  return res;
}

exports.TsPrinter = Printer;
exports.camelCase = camelCase;
exports.docBuilders = docBuilders;
exports.printDocToString = printDocToString;
exports.stringWidth = stringWidth;
