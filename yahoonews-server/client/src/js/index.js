var React = require('react');
var ViewApp = require('./View/App.js');

window.onload = function() {
  React.render(React.createElement(ViewApp, null),
               document.body);
};
