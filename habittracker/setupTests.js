// src/setupTests.js
require('@testing-library/jest-dom');

// Полифил для TextEncoder/TextDecoder
const { TextEncoder, TextDecoder } = require('util');

global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;
