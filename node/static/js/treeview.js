'use strict';

//@include js/common.js
//@include js/api.js

(function () {

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
   elem_append_text(this.ui.name, name);

   this.name = name;
   this.path = path;
   this.children = {};
   this.state = 'none'; // none -> loading -> loaded/error
}
FolderNode.prototype = {
   dom: function () { return this.ui.self; },
   isFolded: function () { return this.ui.items.style.display === 'none'; },
   childrenReset: function () { empty_elem(this.ui.items); },
   childrenLoading: function() {
      var div = document.createElement('div');
      var span = document.createElement('span');
      span.className = 'spin spin-sm';
      div.appendChild(span);
      elem_append_text(div, ' Loading ...');
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
      elem_append_text(div, err || 'Unknown Error');
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
      elem_append_text(div, name);
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
function FolderTree(dom, opt) {
   this.self = dom;
   this.opt = opt || {};
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
               if (that.opt.switchToBrowseTab) {
                  that.opt.switchToBrowseTab();
               }
               if (lastNode) {
                  var name = path.split('/').pop(), ch = null;
                  if (name) {
                     ch = (lastNode.children[name] || {}).dom;
                  } else {
                     ch = lastNode.dom();
                  }
                  if (ch && that.opt.highlightItem) {
                     that.opt.highlightItem(ch);
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
               node.asyncUnfold().then(function () {
                  runLoadingTasks();
               });
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

if (!window.Flame) window.Flame = {};
window.Flame.component = Object.assign(window.Flame.component || {}, {
   FolderNode: FolderNode,
   FolderTree: FolderTree
});

})();
