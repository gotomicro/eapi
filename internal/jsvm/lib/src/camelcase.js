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
export default function camelCase(input, options) {
  let res = '';
  let cursor = 0;
  let wordStart = !!options?.pascalCase;

  while (cursor < input.length) {
    const subStr = input.substr(cursor)
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
        wordStart = true
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
