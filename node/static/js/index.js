'use strict';
//@include js/common.js
//@include js/api.js

(function () {

var ui = {
   btn: {
      nav: {
         search: dom('#btn-search'),
         browse: dom('#btn-browse'),
         team: dom('#btn-team'),
         settings: dom('#btn-settings')
      },
      search: {
         act: dom('#search_btn-search')
      }
   }, // btn
   txt: {
      search: {
         query: dom('#search_txt-search')
      }
   }, // txt
   panel: {
      side: dom('#pnl-side'),
      search: dom('#pnl-side-search'),
      search_result: dom('#search_pnl-item'),
      browse: dom('#pnl-side-browse'),
      browse_tree: dom('#browse_pnl-tree'),
      team: dom('#pnl-side-team'),
      settings: dom('#pnl-side-settings'),
      title: dom('#pnl-title'), // e.g. show breadcrumb
      contents: dom('#pnl-contents'), // e.g. show source code
      mask: dom('.mask')
   }, // panel
   label: {
      username: dom('#lbl-username')
   }, // label
   state: {
      show: function (el) { el.style.display = 'block'; },
      hide: function (el) { el.style.display = 'none'; },
      empty: function (el) { empty_elem(el); },
      flash: function (el, count) {
         var sw = true;
         count = count || 5;
         _flash(el);

         function _flash(el) {
            if (count <= 0) return;
            if (sw) {
               el.classList.remove('item-invert');
            } else {
               el.classList.add('item-invert');
            }
            sw = !sw;
            count --;
            setTimeout(_flash, 200, el);
         }
      },
      append_text: function (el, text) {
         el.appendChild(document.createTextNode(text));
         return el;
      },
      div_message: function (msg, color) {
         var div = document.createElement('div');
         div.className = 'item item-' + (color || 'red');
         div.appendChild(document.createTextNode(msg));
         return div;
      },
      nav: {
         selected: null,
         select: function (el) { el.classList.add('active'); },
         deselect: function (el) { el.classList.remove('active'); }
      },
      treeview: {
         scroll_to: function (el) {
            if (ui.state.nav.selected !== 'browse') return;
            var top = el.offsetTop - ui.panel.side.offsetTop;
            var top0 = ui.panel.side.scrollTop;
            var h = el.offsetHeight;
            var h0 = ui.panel.side.offsetHeight;
            var x = ui.panel.side.scrollLeft;
            if (top0 > top) {
               ui.panel.side.scrollTo(x, top);
            } else if (top0 + h0 - h < top) {
               var y = top - h0 + h;
               if (y < 0) y = 0;
               ui.panel.side.scrollTo(x, y);
            }
         }
      },
      label: {
         text: function (el, val) {
            el.innerHTML = '';
            el.appendChild(document.createTextNode(val));
         }
      }, // label
      global: {
         editor: null
      },
      const: {
         pre_font: null
      }
   }
};

function parseHash() {
   var parts = location.hash.split('#');
   parts.shift();
   var obj = {};
   obj.path = parts[0];
   return obj;
}

/*
<div class="item-thin">
   <a class="folder-fold-btn"></a><span>{item-name}</span>
   <div class="item-container folder">
   </div>
</div>
 */
function FolderNode(name, path) {
   this.ui = {
      self: document.createElement('div'),
      fold: document.createElement('a'),
      name: document.createElement('span'),
      items: document.createElement('div')
   }
   this.ui.self.className = 'item-thin';
   this.ui.fold.className = 'folder-fold-btn';
   this.ui.fold.setAttribute('data-path', path);
   this.ui.items.className = 'item-container folder';
   this.ui.items.style.display = 'none';
   this.ui.self.appendChild(this.ui.fold);
   this.ui.self.appendChild(this.ui.name);
   this.ui.self.appendChild(this.ui.items);
   ui.state.append_text(this.ui.name, name);

   this.name = name;
   this.path = path;
   this.children = {};
   this.state = 'none'; // none -> loading -> loaded/error
}
FolderNode.prototype = {
   dom: function () { return this.ui.self; },
   isFolded: function () { return this.ui.items.style.display === 'none'; },
   childrenReset: function () { ui.state.empty(this.ui.items); },
   childrenLoading: function() {
      var div = document.createElement('div');
      var span = document.createElement('span');
      span.className = 'spin spin-sm';
      div.appendChild(span);
      ui.state.append_text(div, ' Loading ...');
      this.childrenReset();
      this.ui.items.appendChild(div);
      return div;
   },
   childrenError: function (err) {
      var div = document.createElement('div');
      div.className = 'item-thin item-red';
      var img = document.createElement('img');
      img.src = 'img/exclamation.svg';
      img.style.width = '14px';
      img.style.height = '14px';
      div.appendChild(img);
      ui.state.append_text(div, err || 'Unknown Error');
      this.childrenReset();
      this.ui.items.appendChild(div);
      return div;
   },
   addItem: function (node) {
      this.ui.items.appendChild(node.dom());
      return node.dom();
   },
   addFileItem: function (name, url) {
      var div = document.createElement('a');
      div.className = 'item-thin';
      div.href = url;
      ui.state.append_text(div, name);
      this.ui.items.appendChild(div);
      return div;
   },
   asyncFold: function () {
      this.ui.items.style.display = 'none';
      this.ui.fold.classList.remove('active');
      return Promise.resolve(true);
   },
   asyncUnfold: function () {
      this.ui.fold.classList.add('active');
      this.ui.items.style.display = 'block';
      if (this.state !== 'none' && this.state !== 'error') {
         // loading, loaded
         return Promise.resolve(true);
      }
      var that = this;
      this.state = 'loading';
      this.childrenLoading();
      return new Promise(function(r, e) {
         Flame.api.project.getDirectoryContents(that.path).then(
            function (result) {
               that.childrenReset();
               if (!result || !result.length) {
                  that.state = 'error';
                  return r(false);
               }
               result.forEach(function (item) {
                  if (!item.name) return;
                  if (item.name.endsWith('/')) {
                     var node = new FolderNode(
                        item.name.substring(0, item.name.length-1),
                        that.path + item.name
                     );
                     that.addItem(node);
                     that.children[item.name] = node;
                  } else {
                     var div = that.addFileItem(item.name, '#' + that.path + item.name);
                     that.children[item.name] = {
                        file: true,
                        dom: div
                     };
                  }
               });
               that.state = 'loaded';
               r(true);
            },
            function (err) {
               that.childrenError(err);
               that.state = 'error';
               r(false);
            }
         );
      });
   }
};
function FolderTree(dom) {
   this.self = dom;
   // none -> loading -> loaded
   //               \--> error
   this.root = new FolderNode('', '/');
   this.root.ui.self = this.self;
   this.root.ui.items = this.self;
   this.root.childrenReset();
   this.task_queue = [];

   var that = this;
   this.root.ui.self.addEventListener('click', function (evt) {
      var el = evt.target;
      if (el.classList.contains('folder-fold-btn')) {
         var node = that.locateNode(el.getAttribute('data-path'));
         if (!node) return;
         if (node.isFolded()) {
            node.asyncUnfold();
         } else {
            node.asyncFold();
         }
      }
   })
}
FolderTree.prototype = {
   dom: function () { return this.self; },
   locateNode: function (path) {
      if (!path || !path.endsWith('/')) return null;
      var node = this.root;
      var parts = path.split('/');
      // /path/to/folder/ -> '', 'path', 'to', 'folder', ''
      parts.shift();
      parts.pop();
      while (node && parts.length) {
         var name = parts.shift();
         node = node.children[name + '/'];
      }
      return node;
   },
   asyncExpandTo: function (path) {
      var that = this;
      return new Promise(function (r, e) {
         var lastNode = null;
         waitForRootLoaded();

         function waitForRootLoaded() {
            if (that.root.state === 'loaded') {
               generateLoadingTasks();
               return;
            } else if (that.root.state === 'error') {
               return e('failed to get project list.');
            }
            setTimeout(waitForRootLoaded, 100);
         }

         function generateLoadingTasks() {
            that.task_queue = [];
            var parts = path.split('/');
            // e.g. /path/to/folder/ -> path, to, folder
            //      /path/to/file -> path, to
            parts.shift();
            parts.pop();
            parts.forEach(function (name) {
               var task = that.task_queue[that.task_queue.length-1] || { path: '/' };
               that.task_queue.push({ base: task.path, path: task.path + name + '/', name: name + '/' });
            });
            runLoadingTasks();
         }

         function runLoadingTasks() {
            if (!that.task_queue.length) {
               if (ui.state.nav.selected !== 'browse') {
                  onSwitchSidePanel({ target: ui.btn.nav.browse });
               }
               if (lastNode) {
                  var name = path.split('/').pop(), ch = null;
                  if (name) {
                     ch = (lastNode.children[name] || {}).dom;
                  } else {
                     ch = lastNode.dom();
                  }
                  if (ch) {
                     ui.state.treeview.scroll_to(ch);
                     ui.state.flash(ch);
                  }
                  lastNode = null;
               }
               return r();
            }
            runLoadingTask(that.task_queue.shift());
         }

         function runLoadingTask(task) {
            var node = that.locateNode(task.base);
            if (!node) return e('cannot load: ' + task.path);
            node = node.children[task.name];
            if (!node) return e('not found: ' + task.name + ' in ' + task.base);
            lastNode = node;
            if (node.state === 'loaded') {
               runLoadingTasks();
            } else if (node.state === 'loading') {
               waitForNodeLoaded(node);
            } else {
               node.asyncUnfold().then(
                  function (ok) { if (ok) runLoadingTasks(); else e('cannot load: ' + task.path); },
                  function () { e('cannot load: ' + task.path); }
               );
            }

            function waitForNodeLoaded(node) {
               if (node.state === 'loaded') {
                  runLoadingTasks();
               } else if (node.state === 'error') {
                  return e('failed: ' + node.path);
               }
               setTimeout(waitForNodeLoaded, 100, node);
            }
         }
      });
   }
};

function onHashChange() {
   console.log(location.hash);
   var obj = parseHash();
   if (obj.path) {
      if (obj.path.startsWith('/')) {
         ui.panel.browse_tree.asyncExpandTo(obj.path);
         if (ui.state.global.editor) ui.state.global.editor.dispose();
         ui.state.empty(ui.panel.contents);
         ui.state.empty(ui.panel.title);
         renderBreadcrumb(obj.path);
         if (!obj.path.endsWith('/')) {
            onView(obj.path);
         }
      } else if (obj.path.startsWith('?')) {
         var query = decodeURIComponent(obj.path.substring(1));
         onSearch(query);
         if (ui.state.nav.selected !== 'search') {
            onSwitchSidePanel({ target: ui.btn.nav.search });
         }
      }
   }
}

function onSwitchSidePanel(evt) {
   var id = evt.target.getAttribute('id').split('-')[1];
   if (!id) return;
   if (ui.state.nav.selected === id) {
      ui.state.nav.selected = null;
      ui.state.hide(ui.panel.side);
      ui.state.nav.deselect(ui.btn.nav[id]);
   } else {
      ui.state.hide(ui.panel.search);
      ui.state.hide(ui.panel.browse);
      ui.state.hide(ui.panel.team);
      ui.state.hide(ui.panel.settings);
      ui.state.nav.deselect(ui.btn.nav.search);
      ui.state.nav.deselect(ui.btn.nav.browse);
      ui.state.nav.deselect(ui.btn.nav.team);
      ui.state.nav.deselect(ui.btn.nav.settings);
      ui.state.nav.select(ui.btn.nav[id]);
      ui.state.show(ui.panel[id]);
      ui.state.show(ui.panel.side);
      ui.state.nav.selected = id;
   }
}

/*
<div class="item-thin item-blue">
   <a>{path}</a>
   <div class="flex-table flex-row search-match-item item-thin item-purple">
      <div style="font: 13px monospace;"><a>{line-number}</a></div>
      <pre class="flex-auto"><a>{matched-text}</a></pre>
   </div>
</div>
 */
function renderSearchItem(item, opt) {
   var div = document.createElement('div');
   div.className = 'item-thin item-blue search-item';
   var a = document.createElement('a');
   ui.state.append_text(a, item.path);
   a.href = '#' + item.path;
   if (!item.matches || !item.matches.length) {
      return div;
   }
   var match = document.createElement('div');
   match.className = 'search-match-item item-thin item-purple';
   match.style.width = '100%';
   match.style.overflowX = 'auto';
   var lineno = document.createElement('div');
   lineno.style.font = ui.state.const.pre_font;
   lineno.style.float = 'left';
   lineno.style.position = 'sticky';
   lineno.style.left = '0px';
   lineno.style.backgroundColor = 'white';
   var pre = document.createElement('pre');
   pre.className = 'flex-auto';
   item.matches.forEach(function (match, i) {
      if (i > 0) {
         lineno.appendChild(document.createElement('br'));
         pre.appendChild(document.createElement('br'));
      }
      var a = document.createElement('a');
      var span = document.createElement('a');
      ui.state.append_text(a, item.L);
      ui.state.append_text(span, item.T);
      span.href = '#' + item.path + '#L' + match.L;
      lineno.appendChild(a);
      pre.appendChild(span);
   });
   div.appendChild(a);
   match.appendChild(lineno);
   match.appendChild(pre);
   div.appendChild(match);
   return div;
}
function renderSearchItems(result) {
   ui.state.empty(ui.panel.search_result);
   if (!result.items || !result.items.length) {
      ui.panel.search_result.appendChild(ui.state.div_message('Search result: nothing found.'));
      return;
   }
   var opt = {};
   // TODO: add options to opt object, e.g. opt.regexp = result.matchRegexp
   result.items.forEach(function (item) {
      ui.panel.search_result.appendChild(renderSearchItem(item, opt));
   });
}

function onView(path, opt) {
   if (!opt) opt = {};
   // TODO: locate to sepcified line number
   if (ui.state.global.editor) ui.state.global.editor.dispose();
   ui.state.empty(ui.panel.contents);
   Flame.api.project.getFileContents(path).then(
      function (obj) {
         var editor = new Flame.editor.SourceCodeViewer(ui.panel.contents, obj.data);
         ui.state.global.editor = editor;
      },
      function () {}
   );
}

function renderBreadcrumb(path) {
   var parts = path.split('/');
   parts.shift();
   var last = parts.pop();
   if (!last) last = parts.pop();
   if (!last) return;
   ui.state.append_text(ui.panel.title, '# / ');
   var curpath = '/';
   parts.forEach(function (name) {
      curpath += name + '/';
      var a = document.createElement('a');
      ui.state.append_text(a, name);
      a.href = '#' + curpath;
      ui.panel.title.appendChild(a);
      ui.state.append_text(ui.panel.title, ' / ')
   });
   ui.state.append_text(ui.panel.title, last);
}

function onSearch(query) {
   // TODO: handle prev query processing, e.g. cancel, parallel
   ui.txt.search.query.value = query;
   ui.state.empty(ui.panel.search_result);
   // TODO: convert to code instead of html string
   ui.panel.search_result.innerHTML = '<div><span class="spin spin-sm"></span> Searching ...</div>';
   Flame.api.project.search(query).then(
      function (result) {
         renderSearchItems(result);
      },
      function () {
         // TODO: handle errors
      }
   );
}

function onSearchFromBtn(evt) {
   var query = ui.txt.search.query.value;
   if (!query) {
      ui.txt.search.query.focus();
      return;
   }
   location.hash = '#?' + encodeURIComponent(query);
   // hashchange -> onSearch(query);
}

function onSearchFromInput(evt) {
   if (evt.code !== 'Enter') return;
   var query = ui.txt.search.query.value;
   if (!query) return;
   location.hash = '#?' + encodeURIComponent(query);
   // hashchange -> onSearch(query);
}

function initEvent() {
   window.addEventListener('hashchange', onHashChange);
   ui.btn.nav.search.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.browse.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.team.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.settings.addEventListener('click', onSwitchSidePanel);
   ui.btn.search.act.addEventListener('click', onSearchFromBtn);
   ui.txt.search.query.addEventListener('keypress', onSearchFromInput)
}

function initComponent() {
   var cookie = get_cookie();
   ui.state.label.text(ui.label.username, cookie.user || '(flame)');
   // TODO: if cookie.user is empty

   ui.panel.browse_tree = new FolderTree(ui.panel.browse_tree);
   ui.panel.browse_tree.root.asyncUnfold();
   onHashChange();

   var pre = document.createElement('pre');
   document.body.appendChild(pre);
   ui.state.const.pre_font = getComputedStyle(pre).font;
   document.body.removeChild(pre);
}

function initApplication() {
   Flame.api.user.checkLogin().then(
      function () {
         initComponent();
         initEvent();
      },
      function () {
         // if 401, redirect to login page
         // otherwise, message server error
      }
   );
}

initApplication();

})();
