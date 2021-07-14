'use strict';

//@include css/editor.css
//@include js/common.js

(function () {

function SourceCodeViewer(dom, text) {
   this.lines = text.split('\n');

   this.ui = {
      self: dom,
      leftSide: document.createElement('div'),
      lineNumber: document.createElement('div'),
      blame: document.createElement('div'),
      text: document.createElement('pre')
   };
   var root = document.createElement('div');
   var sideFlex = document.createElement('div');
   root.className = 'editor-container';
   this.render();
   sideFlex.className = 'editor-side-flex';
   this.ui.leftSide.className = 'editor-linenumber';
   this.ui.text.className = 'editor-text flex-auto';
   sideFlex.appendChild(this.ui.lineNumber);
   this.ui.blame.style.display = 'none';
   sideFlex.appendChild(this.ui.blame);
   this.ui.leftSide.appendChild(sideFlex);
   root.appendChild(this.ui.leftSide);
   root.appendChild(this.ui.text);
   empty_elem(this.ui.self);
   this.ui.self.appendChild(root);
   sideFlex.style.font = getComputedStyle(this.ui.text).font;
}
SourceCodeViewer.prototype = {
   render: function () {
      // TODO: show blame info for each line in this.ui.blame (by default display: none)
      // TODO: syntax highlight for source code text
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
