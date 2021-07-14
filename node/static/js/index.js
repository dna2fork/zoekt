'use strict';
//@include js/common.js
//@include js/api.js

(function () {

var ui = {
   btn: {
      nav: {
         search: dom('#btn-search'),
         browse: dom('#btn-browse'),
         analysis: dom('#btn-analysis'),
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
      analysis: dom('#pnl-side-analysis'),
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
      flash: function (el) { elem_flash(el); },
      append_text: function (el, text) { elem_append_text(el, text); },
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

function onHashChange() {
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
      var all = ['search', 'browse', 'analysis', 'team', 'settings'];
      all.forEach(function (one) {
         ui.state.hide(ui.panel[one]);
         ui.state.nav.deselect(ui.btn.nav[one]);
      });
      ui.state.nav.select(ui.btn.nav[id]);
      ui.state.show(ui.panel[id]);
      ui.state.show(ui.panel.side);
      ui.state.nav.selected = id;
   }
}

/*
<div class="item-thin item-blue">
   <a>{path}</a>
   <div class="flex-table flex-row item-thin item-purple">
      <div class="editor-linenumber" style="font: 13px monospace;"><a>{line-number}</a></div>
      <pre class="editor-text flex-auto"><a>{matched-text}</a></pre>
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
   match.className = 'item-thin item-purple';
   match.style.width = '100%';
   match.style.overflowX = 'auto';
   var lineno = document.createElement('div');
   lineno.className = 'editor-linenumber';
   lineno.style.font = ui.state.const.pre_font;
   var pre = document.createElement('pre');
   pre.className = 'editor-text flex-auto';
   item.matches.forEach(function (match, i) {
      if (i > 0) {
         lineno.appendChild(document.createElement('br'));
         pre.appendChild(document.createElement('br'));
      }
      var a = document.createElement('a');
      var span = document.createElement('a');
      ui.state.append_text(a, match.L);
      ui.state.append_text(span, match.T);
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

function renderNotSupportFileView(obj) {
   ui.state.empty(ui.panel.contents);
   var div = document.createElement('div');
   div.className = 'item item-red';
   var img = document.createElement('img');
   img.src = 'img/exclamation.svg';
   img.style.width = '18px';
   img.style.height = '18px';
   div.appendChild(img);
   ui.state.append_text(div, ' binary file view NOT supported yet!');
   ui.panel.contents.appendChild(div);
   return { dispose: function () {} };
}

function onView(path, opt) {
   if (!opt) opt = {};
   // TODO: locate to sepcified line number
   if (ui.state.global.editor) ui.state.global.editor.dispose();
   ui.state.empty(ui.panel.contents);
   Flame.api.project.getFileContents(path).then(
      function (obj) {
         var editor;
         if (obj.binary) {
            editor = renderNotSupportFileView(obj);
         } else {
            editor = new Flame.editor.SourceCodeViewer(ui.panel.contents, obj.data);
         }
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
      a.setAttribute('data-path', curpath);
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

function onClickBreadcrumb(evt) {
   if (evt.target.tagName.toLowerCase() !== 'a') return;
   if (!evt.ctrlKey && !evt.shiftKey && !evt.altKey) {
      evt.preventDefault();
   }
   var path = evt.target.getAttribute('data-path');
   if (!path) return;
   ui.panel.browse_tree.asyncExpandTo(path);
}

function initEvent() {
   window.addEventListener('hashchange', onHashChange);
   ui.btn.nav.search.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.browse.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.analysis.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.team.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.settings.addEventListener('click', onSwitchSidePanel);
   ui.btn.search.act.addEventListener('click', onSearchFromBtn);
   ui.txt.search.query.addEventListener('keypress', onSearchFromInput);
   ui.panel.title.addEventListener('click', onClickBreadcrumb);
}

function initComponent() {
   var cookie = get_cookie();
   ui.state.label.text(ui.label.username, cookie.user || '(flame)');
   // TODO: if cookie.user is empty

   ui.panel.browse_tree = new Flame.component.FolderTree(ui.panel.browse_tree, {
      switchToBrowseTab: function () {
         if (ui.state.nav.selected !== 'browse') {
            onSwitchSidePanel({ target: ui.btn.nav.browse });
         }
      },
      highlightItem: function (elem) {
         ui.state.treeview.scroll_to(elem);
         ui.state.flash(elem);
      }
   });
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
