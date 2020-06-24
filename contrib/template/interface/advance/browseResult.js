function zoektBrowse(params) {
   var cookie = zoektGetCookie();
   var auth = cookie.zoekt_auth?('Basic ' + cookie.zoekt_auth):'';
   return zoektAjax({
      method: 'GET',
      url: '/fsprint',
      data: params,
      headers: { 'Authorization': auth }
   });
}

function zoektBrowseResultRender(elem, obj) {
   elem.innerHTML = '';
   if (obj.directory) {
      zoektBrowseResultRenderForFolder(elem, obj);
   } else {
      zoektBrowseResultRenderForFile(elem, obj);
   }
}

function zoektBrowseResultRenderForFolder(elem, obj) {
   var div = document.createElement('div'), a;
   div.appendChild(document.createTextNode('@' + obj.meta.repo + ':' + obj.meta.path + ' '));
   a = document.createElement('a');
   a.innerHTML = '..';
   a.href = 'javascript:void(0)';
   a.className = 'browse-folder';
   a.setAttribute('data-repo', obj.meta.repo);
   var updir = obj.meta.path.split('/');
   updir.pop(); updir.pop();
   updir = updir.join('/');
   a.setAttribute('data-path', updir?(updir + '/'):'/');
   div.appendChild(a);
   elem.appendChild(div);
   obj.contents.forEach(function (item) {
      if (!item) return;
      div = document.createElement('div');
      a = document.createElement('a');
      a.appendChild(document.createTextNode(item.name));
      a.href = 'javascript:void(0)';
      if (item.name.charAt(item.name - 1) === '/') {
         a.className = 'browse-folder';
      } else {
         a.className = 'browse-file';
      }
      a.setAttribute('data-repo', obj.meta.repo);
      a.setAttribute('data-path', obj.meta.path + item.name);
      div.appendChild(a);
      elem.appendChild(div);
   });
}

function zoektBrowseResultRenderForFile(elem, obj) {
   var div = document.createElement('div'), a;
   div.appendChild(document.createTextNode('@' + obj.meta.repo + ':' + obj.meta.path + ' '));
   a = document.createElement('a');
   a.innerHTML = '..';
   a.href = 'javascript:void(0)';
   a.className = 'browse-folder';
   a.setAttribute('data-repo', obj.meta.repo);
   var updir = obj.meta.path.split('/');
   updir.pop();
   updir = updir.join('/');
   a.setAttribute('data-path', updir?(updir + '/'):'/');
   div.appendChild(a);
   elem.appendChild(div);
   var pre = document.createElement('pre');
   zoektFileRender(pre, obj.contents);
   elem.appendChild(pre);
}

function zoektFileRender(elem, text) {
   var div, span;
   text.split('\n').forEach(function (line) {
      div = document.createElement('div');
      div.appendChild(document.createTextNode(line));
      elem.appendChild(div);
   });
}

function zoektBrowseEvents() {
   var div = document.querySelector('#result');
   div.addEventListener('click', function (evt) {
      if (evt.target.classList.contains('browse-folder') || evt.target.classList.contains('browse-file')) {
         var repo = evt.target.getAttribute('data-repo');
         var path = evt.target.getAttribute('data-path');
         // browseResult.js#zoektBrowse
         zoektBrowse({ a: 'get', r: repo, f: path }).then(function (xhr) {
            var input = document.querySelector('#txt_query');
            input.value = 'r:' + repo + ' f:' + path
            try {
               var obj = JSON.parse(xhr.response);
               if (obj.error) throw 'error';
               obj.meta = { repo: repo, path: path };
               // browseResult.js#zoektBrowseResultRender
               zoektBrowseResultRender(div, obj);
            } catch (e) {
               div.innerHTML = '(internal error)';
            }
         });
      }
   });
}

zoektBrowseEvents();
