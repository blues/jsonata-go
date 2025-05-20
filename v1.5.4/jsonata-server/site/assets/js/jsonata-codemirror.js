
// CodeMirror syntax highlighting rules for JSONata.

CodeMirror.defineMode("jsonata", function(config, parserConfig) {
    var templateMode = parserConfig.template;
    var jsonata = parserConfig.jsonata;

    const operators = {
        '.': 75,
        '[': 80,
        ']': 0,
        '{': 70,
        '}': 0,
        '(': 80,
        ')': 0,
        ',': 0,
        '@': 75,
        '#': 70,
        ';': 80,
        ':': 80,
        '?': 20,
        '+': 50,
        '-': 50,
        '*': 60,
        '/': 60,
        '%': 60,
        '|': 20,
        '=': 40,
        '<': 40,
        '>': 40,
        '`': 80,
        '**': 60,
        '..': 20,
        ':=': 30,
        '!=': 40,
        '<=': 40,
        '>=': 40,
        'and': 30,
        'or': 25,
        '||' : 50,
        '!': 0   // not an operator, but needed as a stop character for name tokens
    };

    const escapes = {  // JSON string escape sequences - see json.org
        '"': '"',
        '\\': '\\',
        '/': '/',
        'b': '\b',
        'f': '\f',
        'n': '\n',
        'r': '\r',
        't': '\t'
    };

    var tokenizer = function(path) {
        var position = 0;
        var length = path.length;

        var create = function(type, value) {
            var obj = { type: type, value: value, position: position};
            return obj;
        };

        var next = function() {
            if(position >= length) return null;
            var currentChar = path.charAt(position);
            // skip whitespace
            while(position < length && ' \t\n\r\v'.indexOf(currentChar) > -1) {
                position++;
                currentChar = path.charAt(position);
            }
            // handle double-char operators
            if(currentChar === '.' && path.charAt(position+1) === '.') {
                // double-dot .. range operator
                position += 2;
                return create('operator', '..');
            }
            if(currentChar === '|' && path.charAt(position+1) === '|') {
                // double-pipe || string concatenator
                position += 2;
                return create('operator', '||');
            }
            if(currentChar === ':' && path.charAt(position+1) === '=') {
                // := assignment
                position += 2;
                return create('operator', ':=');
            }
            if(currentChar === '!' && path.charAt(position+1) === '=') {
                // !=
                position += 2;
                return create('operator', '!=');
            }
            if(currentChar === '>' && path.charAt(position+1) === '=') {
                // >=
                position += 2;
                return create('operator', '>=');
            }
            if(currentChar === '<' && path.charAt(position+1) === '=') {
                // <=
                position += 2;
                return create('operator', '<=');
            }
            if(currentChar === '*' && path.charAt(position+1) === '*') {
                // **  descendant wildcard
                position += 2;
                return create('operator', '**');
            }
            // test for operators
            if(operators.hasOwnProperty(currentChar)) {
                position++;
                return create('operator', currentChar);
            }
            // test for string literals
            if(currentChar === '"' || currentChar === "'") {
                var quoteType = currentChar;
                // double quoted string literal - find end of string
                position++;
                var qstr = "";
                while(position < length) {
                    currentChar = path.charAt(position);
                    if(currentChar === '\\') { // escape sequence
                        position++;
                        currentChar = path.charAt(position);
                        if(escapes.hasOwnProperty(currentChar)) {
                            qstr += escapes[currentChar];
                        } else if(currentChar === 'u') {
                            // \u should be followed by 4 hex digits
                            var octets = path.substr(position+1, 4);
                            if(/^[0-9a-fA-F]+$/.test(octets)) {
                                var codepoint = parseInt(octets, 16);
                                qstr += String.fromCharCode(codepoint);
                                position += 4;
                            } else {
                                throw new Error('The escape sequence \\u must be followed by 4 hex digits at column ' + position);
                            }
                        } else {
                            // illegal escape sequence
                            throw new Error('unsupported escape sequence: \\' + currentChar + ' at column ' + position);
                        }
                    } else if(currentChar === quoteType) {
                        position++;
                        return create('string', qstr);
                    } else {
                        qstr += currentChar;
                    }
                    position++;
                }
                throw new Error('no terminating quote found in string literal starting at column ' + position);
            }
            // test for numbers
            var numregex = /^-?(0|([1-9][0-9]*))(\.[0-9]+)?([Ee][-+]?[0-9]+)?/;
            var match = numregex.exec(path.substring(position));
            if(match !== null) {
                var num = parseFloat(match[0]);
                if(!isNaN(num) && isFinite(num)) {
                    position += match[0].length;
                    return create('number', num);
                } else {
                    throw new Error('Number out of range: ' + match[0]);
                }
            }
            // test for names
            var i = position;
            var ch;
            var name;
            while(true) {
                ch = path.charAt(i);
                if(i == length || ' \t\n\r\v'.indexOf(ch) > -1 || operators.hasOwnProperty(ch)) {
                    if(path.charAt(position) === '$') {
                        // variable reference
                        name = path.substring(position + 1, i);
                        position = i;
                        return create('variable', name);
                    } else {
                        name = path.substring(position, i);
                        position = i;
                        switch(name) {
                            case 'and':
                            case 'or':
                                return create('operator', name);
                            case 'true':
                                return create('value', true);
                            case 'false':
                                return create('value', false);
                            case 'null':
                                return create('value', null);
                            default:
                                if(position == length && name === '') {
                                    // whitespace at end of input
                                    return null;
                                }
                                return create('name', name);
                        }
                    }
                } else {
                    i++;
                }
            }
        };

        return next;
    };

    var templatizer = function(text) {
        var position = 0;
        var length = text.length;

        var create = function(type, value) {
            var obj = { type: type, value: value, position: position};
            return obj;
        };

        var next = function() {
            if(position >= length) return null;
            var currentChar = text.charAt(position);
            // skip whitespace
            while(position < length && ' \t\n\r\v'.indexOf(currentChar) > -1) {
                position++;
                currentChar = text.charAt(position);
            }

            if(currentChar === '{' && text.charAt(position+1) === '{') {
                // found {{
                position += 2;
                // parse what follows using the jsonata parser
                var rest = text.substring(position);
                try {
                    jsonata.parser(rest);
                    // if we get here, we parsed to the end of the buffer with no closing handlebars
                    position += rest.length;
                    return create('variable');
                } catch (err) {
                    if (err.token === '(end)') {
                        position = length;
                        return create('variable');
                    }
                    if (rest.charAt(err.position - 1) != "}" || rest.charAt(err.position) != "}") {
                        // no closing handlbars
                        position += err.position;
                        return create('variable');
                    }
                    position += err.position + 1;
                    return create('variable');
                }
            } else {
                // search forward for next {{
                position = text.indexOf("{{", position);
                if(position != -1) {
                    return create('operator');
                }
                position = length;
                return create('operator');
            }

        };

        return next;
    };

    var TOKEN_NAMES = {
        'operator': 'operator',
        'variable': 'string-2',
        'string': 'string',
        'number': 'number',
        'value': 'keyword',
        'name': 'attribute'
    };

    var currentIndent = 0;

    return {
        token: function(stream) {
            var lexer;
            if(templateMode) {
                lexer = templatizer(stream.string.substr(stream.pos));
            } else {
                lexer = tokenizer(stream.string.substr(stream.pos));
            }
            var token;
            try {
                token = lexer();
            } catch(err) {
                token = null;
            }
            if(token === null) {
                stream.skipToEnd();
                return null;
            }
            var length = token.position;
            while(length > 0) {
                stream.next();
                length--;
            }

            var style = TOKEN_NAMES[token.type];
            return style;
        }
    };
});
