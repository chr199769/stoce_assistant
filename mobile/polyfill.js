const util = require('util');

if (!URL.canParse) {
  URL.canParse = function (url, base) {
    try {
      new URL(url, base);
      return true;
    } catch (e) {
      return false;
    }
  };
}

if (!util.styleText) {
  // Simple polyfill that ignores the style and returns the text
  // The format argument is ignored
  util.styleText = function (format, text) {
    return text;
  };
}

if (!Array.prototype.toReversed) {
  Array.prototype.toReversed = function () {
    return this.slice().reverse();
  };
}

if (!Array.prototype.toSorted) {
  Array.prototype.toSorted = function (compareFn) {
    return this.slice().sort(compareFn);
  };
}

if (!Array.prototype.toSpliced) {
  Array.prototype.toSpliced = function (start, deleteCount, ...items) {
    const copy = this.slice();
    copy.splice(start, deleteCount, ...items);
    return copy;
  };
}

if (!Array.prototype.with) {
  Array.prototype.with = function (index, value) {
    const copy = this.slice();
    copy[index] = value;
    return copy;
  };
}
