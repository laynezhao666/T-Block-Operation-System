/* eslint-disable */
const pairSplitRegExp = /; */;
const decode = decodeURIComponent;

function tryDecode(str, func) {
    try {
        return func(str);
    } catch (e) {
        return str;
    }
}

/**
 * Parse a cookie header.
 *
 * Parse the given cookie header string into an object
 * The object has the various cookies as keys(names) => values
 *
 * @param {string} str
 * @param {object} [options]
 * @return {object}
 * @public
 */

function parse(str, options) {
    if (typeof str !== 'string') {
        throw new TypeError('argument str must be a string');
    }

    const obj = {};
    const opt = options || {};
    const pairs = str.split(pairSplitRegExp);
    const dec = opt.decode || decode;

    for (let i = 0; i < pairs.length; i++) {
        const pair = pairs[i];
        let eqIndex = pair.indexOf('=');

        // skip things that don't look like key=value
        if (eqIndex < 0) {
            continue;
        }

        const key = pair.substr(0, eqIndex).trim();
        let val = pair.substr(++eqIndex, pair.length).trim();

        // quoted values
        if (val[0] == '"') {
            val = val.slice(1, -1);
        }

        // only assign once
        if (!obj[key]) {
            obj[key] = tryDecode(val, dec);
        }
    }

    return obj;
}


export default parse;
/* eslint-disable */
