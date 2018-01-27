// Imports and uses the 'main.js' generated through gopherjs
'use strict';
console.log('Starting');

console.log('Loading main.js');
require('./main.js'); // Attaches methods to a 'user' property on the 'global' object
var main = global.goTemplateParser;

var template = '<h1>{{ .text }}</h1>';
var data = { text: 'data text'};

console.log(main.compile(data, template));