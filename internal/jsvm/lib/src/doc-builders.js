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
export function assertDoc(val) {
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
export function concat(parts) {
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
export function indent(contents) {
  assertDoc(contents);
  return {type: "indent", contents};
}

/**
 * @param {number | string} widthOrString
 * @param {Doc} contents
 * @returns Doc
 */
export function align(widthOrString, contents) {
  assertDoc(contents);

  return {type: "align", contents, n: widthOrString};
}

/**
 * @param {Doc} contents
 * @param {object} [opts] - TBD ???
 * @returns Doc
 */
export function group(contents, opts = {}) {
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
export function dedentToRoot(contents) {
  return align(Number.NEGATIVE_INFINITY, contents);
}

/**
 * @param {Doc} contents
 * @returns Doc
 */
export function markAsRoot(contents) {
  // @ts-expect-error - TBD ???:
  return align({type: "root"}, contents);
}

/**
 * @param {Doc} contents
 * @returns Doc
 */
export function dedent(contents) {
  return align(-1, contents);
}

/**
 * @param {Doc[]} states
 * @param {object} [opts] - TBD ???
 * @returns Doc
 */
export function conditionalGroup(states, opts) {
  return group(states[0], {...opts, expandedStates: states});
}

/**
 * @param {Doc[]} parts
 * @returns Doc
 */
export function fill(parts) {
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
export function ifBreak(breakContents, flatContents, opts = {}) {
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
export function indentIfBreak(contents, opts) {
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
export function lineSuffix(contents) {
  assertDoc(contents);
  return {type: "line-suffix", contents};
}

export const lineSuffixBoundary = {type: "line-suffix-boundary"};
export const breakParent = {type: "break-parent"};
export const trim = {type: "trim"};

export const hardlineWithoutBreakParent = {type: "line", hard: true};
export const literallineWithoutBreakParent = {
  type: "line",
  hard: true,
  literal: true,
};

export const line = {type: "line"};
export const softline = {type: "line", soft: true};
// eslint-disable-next-line prettier-internal-rules/no-doc-builder-concat
export const hardline = concat([hardlineWithoutBreakParent, breakParent]);
// eslint-disable-next-line prettier-internal-rules/no-doc-builder-concat
export const literalline = concat([literallineWithoutBreakParent, breakParent]);

export const cursor = {type: "cursor", placeholder: Symbol("cursor")};

/**
 * @param {Doc} sep
 * @param {Doc[]} arr
 * @returns Doc
 */
export function join(sep, arr) {
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
export function addAlignmentToDoc(doc, size, tabWidth) {
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

export function label(label, contents) {
  return {type: "label", label, contents};
}
