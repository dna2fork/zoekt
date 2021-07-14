'use strict';

//@include css/editor.css
//@include js/common.js

(function () {

function SourceCodeViewer(dom, text) {
   this.lines = text.split('\n');

   this.ui = {
      self: dom,
      lineNumber: document.createElement('div'),
      text: document.createElement('pre')
   };
   var root = document.createElement('div');
   root.className = 'editor-container';
   this.render();
   this.ui.lineNumber.className = 'editor-linenumber';
   this.ui.text.className = 'editor-text flex-auto';
   root.appendChild(this.ui.lineNumber);
   root.appendChild(this.ui.text);
   empty_elem(this.ui.self);
   this.ui.self.appendChild(root);
   this.ui.lineNumber.style.font = getComputedStyle(this.ui.text).font;
}
SourceCodeViewer.prototype = {
   render: function () {
      empty_elem(this.ui.lineNumber);
      empty_elem(this.ui.text);
      var that = this;
      this.lines.forEach(function (line, i) {
         if (i > 0) {
            that.ui.lineNumber.appendChild(document.createElement('br'));
            that.ui.text.appendChild(document.createElement('br'));
         }
         var a = document.createElement('a');
         a.appendChild(document.createTextNode('' + (i + 1)));
         that.ui.lineNumber.appendChild(a);
         var span = document.createElement('span');
         span.appendChild(document.createTextNode(line));
         that.ui.text.appendChild(span);
      });
   },
   dispose: function () {}
};

if (!window.Flame) window.Flame = {};
window.Flame.editor = {
   SourceCodeViewer: SourceCodeViewer
};

})();
