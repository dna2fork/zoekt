'use strict';

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
   root.style.width = '100%';
   root.style.height = '100%';
   root.style.overflow = 'auto';
   this.render();
   // TODO: move such styles to css class instead of hard code here
   this.ui.lineNumber.style.marginRight = '2px';
   this.ui.lineNumber.style.paddingRight = '2px';
   this.ui.lineNumber.style.borderRight = '1px solid #999';
   this.ui.lineNumber.style.textAlign = 'right';
   this.ui.lineNumber.style.float = 'left';
   this.ui.lineNumber.style.position = 'sticky';
   this.ui.lineNumber.style.left = '0px';
   this.ui.lineNumber.style.backgroundColor = 'white';
   this.ui.text.className = 'flex-auto';
   this.ui.text.style.border = 'none';
   this.ui.text.style.padding = '0px';
   this.ui.text.style.margin = '0px';
   this.ui.text.style.tabSize = '4';
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
