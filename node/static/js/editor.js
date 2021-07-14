'use strict';

//@include css/editor.css
//@include js/common.js
//@include js/onhold.js

(function () {

function SourceInfoBox(dom, opt) {
   // opt:
   //    align --- element for positioning
   this.opt = opt || {};
   this.data = null;
   this.ui = {
      self: dom,
      title: document.createElement('div'),
      close: document.createElement('button'),
   };

   this.ui.self.className = 'editor-info-box';
   this.ui.self.style.display = 'none';

   var div = document.createElement('div');
   div.className = 'flex-table flex-row';
   this.ui.title.className = 'flex-auto';
   this.ui.title.style.padding = '3px';
   this.ui.title.style.margin = '2px';
   this.ui.close.innerHTML = 'X';
   this.ui.close.style.padding = '3px';
   this.ui.close.style.margin = '2px';
   div.appendChild(this.ui.title);
   div.appendChild(this.ui.close);
   this.ui.self.appendChild(div);

   var that = this;
   this.events = {
      onClickClose: function (evt) {
         that.hide();
      }
   };
   this.ui.close.addEventListener('click', this.events.onClickClose);
}
SourceInfoBox.prototype = {
   show: function (data) {
      var y = this.opt.align?this.opt.align.offsetTop:0;
      var h = this.opt.align?this.opt.align.offsetHeight:0;
      this.data = data;
      this.ui.self.style.right = '0px';
      this.ui.self.style.top = y + 'px';
      this.ui.self.style.width = '200px';
      this.ui.self.style.height = h + 'px';
      this.ui.self.style.display = 'block';
   },
   hide: function () {
      this.data = null;
      this.ui.self.style.display = 'none';
   },
   dispose: function () {
      this.ui.close.removeEventListener('click', this.events.onClickClose);
   },
   dom: function () { return this.ui.self; }
};

function SourceCodeViewer(dom, text) {
   this.lines = text.split('\n');

   this.ui = {
      self: dom,
      leftSide: document.createElement('div'),
      lineNumber: document.createElement('div'),
      blame: document.createElement('div'),
      text: document.createElement('pre'),
      highlight: document.createElement('div'),
      info: new SourceInfoBox(
         document.createElement('div'),
         { align: dom }
      )
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
   root.appendChild(this.ui.info.dom());
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
      that.ui.info.hide();
      that.ui.info.show({ linenumber: parseInt(evt.target.textContent, 10) });
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
      this.ui.info.dispose();
   }
};

if (!window.Flame) window.Flame = {};
window.Flame.editor = {
   SourceCodeViewer: SourceCodeViewer
};

})();
