'use strict';

//@include css/editor.css
//@include js/common.js
//@include js/onhold.js

(function () {

function SourceCodeViewer(dom, text) {
   this.lines = text.split('\n');

   this.ui = {
      self: dom,
      leftSide: document.createElement('div'),
      lineNumber: document.createElement('div'),
      blame: document.createElement('div'),
      text: document.createElement('pre'),
      highlight: document.createElement('div')
   };

   var root = document.createElement('div');
   root.className = 'editor-container';
   this.render();

   // post: ui before to screen
   var sideFlex = document.createElement('div');
   sideFlex.className = 'editor-side-flex';
   this.ui.leftSide.className = 'editor-left-side';
   this.ui.text.className = 'editor-text flex-auto';
   this.ui.lineNumber.className = 'editor-linenumber';
   this.ui.highlight.className = 'editor-highlight';
   this.ui.blame.style.display = 'none';

   // assemble components
   sideFlex.appendChild(this.ui.lineNumber);
   sideFlex.appendChild(this.ui.blame);
   this.ui.leftSide.appendChild(sideFlex);
   root.appendChild(this.ui.highlight);
   root.appendChild(this.ui.leftSide);
   root.appendChild(this.ui.text);
   empty_elem(this.ui.self);
   this.ui.self.appendChild(root);
   // post: ui attached to screen
   sideFlex.style.font = getComputedStyle(this.ui.text).font;

   this.onHoldOn = new Flame.event.ElementHoldEvent(this.ui.lineNumber, {
      selectorFn: function (el) { return el.tagName.toLowerCase() === 'a'; }
   });
   var that = this;
   this.onHoldOn.on(function (evt) {
      // TODO: process for long hold click / touch on a line number
      //       show metadata like blame, comment, linkage, ... in 'analysis' tab
   });
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
   dispose: function () {
      this.onHoldOn.dispose();
   }
};

if (!window.Flame) window.Flame = {};
window.Flame.editor = {
   SourceCodeViewer: SourceCodeViewer
};

})();
