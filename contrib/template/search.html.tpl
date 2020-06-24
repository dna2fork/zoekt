<html>
<head>
   <meta charset="utf-8" />
   <meta name="viewport" content="width=device-width, initial-scale=1" />
   <title>Zoekt</title>
   <style>
.header {
   display: flex;
   color: #fafafa;
   background-color: #004a70;
   height: 2.5rem;
   white-space: nowrap;
   margin-bottom: 5px;
}
.header > span {
   margin: auto 0 auto 10px;
}
.header > span > a {
   color: white;
}
a {
   text-decoration: none;
}
.searchbox {
   display: flex;
}
.search-input {
   flex: 1 0 auto;
   padding: 0 2px 0 2px;
}
.search-input > input {
   display: inline-block;
   width: 100%;
   height: 30px;
   font-size: 15px;
   border-top: none;
   border-left: none;
   border-right: none;
   border-bottom: 1px solid #555555;
}
.search-input > input::focus {
   border-bottom: 2px solid black;
}
.search-btn {
   flex: 0 1 auto;
}
.search-btn > button {
   height: 30px;
   font-size: 15px;
   border: 1px solid black;
}
.about {
   margin-top: 10px;
   background-color: #eeeeee;
   border: 1px solid #dddddd;
   padding: 10px;
}
.about > .link {
   padding: 5px 0 5px 0;
}
.result {
   margin-top: 5px;
   margin-bottom: 5px;
}
.code {
   padding: 1px;
   border: 1px solid red;
   background-color: #ffcccc;
}

   </style>
</head>
<body>
<header class="header"><span><a href="#">Zoekt</a></span></header>
<div class="searchbox">
   <div class="search-input"><input id="txt_query" placeholder="Query" /></div>
   <div class="search-btn"><button id="btn_search">Search</button></div>
</div>
<div class="result" id="result"></div>
<div class="about">
   <a class="link" href="about">About</a>
</div>
<script>
function zoektUriencode(obj) {
   var params = [];
   for (var key in obj) {
      params.push(key + '=' + encodeURIComponent(obj[key]));
   }
   return '?' + params.join('&');
}

function zoektAjax(options) {
   return new Promise(function (resolve, reject) {
      var xhr = new XMLHttpRequest(), payload = null;
      xhr.open(options.method || 'POST', options.url + (options.data ? zoektUriencode(options.data) : ''), true);
      xhr.addEventListener('readystatechange', function (evt) {
         if (evt.target.readyState === 4 /*XMLHttpRequest.DONE*/) {
            if (~~(evt.target.status / 100) === 2) {
               resolve(evt.target);
            } else {
               reject(evt.target);
            }
         }
      });
      if (options.headers) {
         for (var header in options.headers) {
            xhr.setRequestHeader(header, options.headers[header]);
         }
      }
      if (options.json) {
         xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
         payload = JSON.stringify(options.json);
      }
      xhr.send(payload);
   });
}

function zoektGetCookie() {
   var items = document.cookie;
   var r = {};
   if (!items) return r;
   items.split(';').forEach(function (one) {
      var p = one.indexOf('=');
      if (p < 0) r[one.trim()] = null;
      else r[one.substring(0, p).trim()] = one.substring(p + 1).trim();
   });
   return r;
}

function zoektSearch(query) {
   var cookie = zoektGetCookie();
   var auth = cookie.zoekt_auth?('Basic ' + cookie.zoekt_auth):'';
   return zoektAjax({
      method: 'GET',
      url: '/search',
      data: { q: query },
      headers: { 'Authorization': auth }
   });
}

function zoektSearchTrigger() {
   var input = document.querySelector('#txt_query');
   var btn = document.querySelector('#btn_search');
   if (!input.value) return;
   btn.setAttribute('disabled', true);
   zoektSearch(input.value).then(function (xhr) {
      btn.removeAttribute('disabled');
      var obj;
      try { obj = JSON.parse(xhr.response); } catch(err) { obj = null; }
      var div = document.querySelector('#result');
      if (obj) {
         if (obj.repositories) {
            obj.repositories.pop();
            obj.repositories.forEach(function (repoObj) { if (repoObj.branches) repoObj.branches.pop(); });
         } else if (obj.hits) {
            obj.hits.pop();
            obj.hits.forEach(function (matchObj) { if (matchObj.matches) matchObj.matches.pop(); });
         }
         var pre = document.createElement('pre');
         // searchResult.js#zoektSearchResultRender
         zoektSearchResultRender(pre, obj);
         div.innerHTML = '';
         div.appendChild(pre);
      } else {
         div.innerHTML = '(No Result)';
      }
   }, function (xhr) {
      btn.removeAttribute('disabled');
      if (xhr.status === 401) {
         var div = document.querySelector('#result');
         div.innerHTML = 'Unauthenticated. Please add cookie "zoekt_auth": <span class="code">document.cookie = "zoekt_auth=" + btoa("username:password")</span>.';
      }
   });
}

function zoektInit() {
   var input = document.querySelector('#txt_query');
   var btn = document.querySelector('#btn_search');
   input.addEventListener('keydown', function (evt) {
      if (evt.keyCode !== 13) return;
      btn.click();
   });
   btn.addEventListener('click', function (evt) {
      zoektSearchTrigger();
   });
}

zoektInit();

function zoektSearchResultRender(pre, obj) {
   // obj.repositories
   // obj.hits
   if (obj.repositories) {
      zoektSearchResultRenderForRepositories(pre, obj)
   } else if (obj.hits) {
      zoektSearchResultRenderForHits(pre, obj)
   }
}

function zoektSearchResultRenderForRepositories(pre, obj) {
   obj.repositories.forEach(function (item) {
      if (!item) return;
      var div = document.createElement('div');
      var a = document.createElement('a');
      a.href = 'javascript:void(0)';
      a.className = 'search-repo';
      a.appendChild(document.createTextNode(item.name));
      div.appendChild(a);
      div.appendChild(document.createTextNode(' (' + item.file_count + ' files, ' + utilBuildSize(item.size) + ')'));
      pre.appendChild(div);
   });
}

function zoektSearchResultRenderForHits(pre, obj) {
   var div, a;
   obj.hits.forEach(function (item) {
      if (!item) return;
      div = document.createElement('div');
      a = document.createElement('a');
      a.href = 'javascript:void(0)';
      a.className = 'browse-file';
      a.setAttribute('data-repo', item.repository);
      a.setAttribute('data-path', '/' + item.filename);
      a.appendChild(document.createTextNode('/' + item.repository + '/' + item.filename));
      div.appendChild(a);
      if (item.matches.length > 1) item.matches.sort(function (a, b) { return a.linenumber - b.linenumber; });
      var N = 0; // max line number string length
      if (item.matches.length) {
         N = item.matches.map(
            function (x) { return ('' + x.linenumber).length; }
         ).reduce(
            function (a, b) { return a>b?a:b }
         );
      }
      item.matches.forEach(function (match) {
         var line = document.createElement('div');
         if (match.linenumber) {
            var span = document.createElement('span');
            span.innerHTML = paddingNumber(match.linenumber, N);
            line.appendChild(span);
            line.innerHTML += ' ' + match.text;
         } else {
            line.innerHTML = match.text;
         }
         div.append(line);
      });
      div.style.marginTop = '5px';
      pre.appendChild(div);
   });

   function paddingNumber(num, N) {
      var str = '' + num;
      while (str.length < N) str = '&nbsp;' + str;
      return str;
   }
}

function zoektSearchEvents() {
   var div = document.querySelector('#result');
   div.addEventListener('click', function (evt) {
      if (evt.target.classList.contains('search-repo')) {
         var repo = evt.target.textContent.trim();
         // browseResult.js#zoektBrowse
         zoektBrowse({ a: 'get', r: repo, f: '/' }).then(function (xhr) {
            var input = document.querySelector('#txt_query');
            input.value = 'r:' + repo + ' f:/'
            try {
               var obj = JSON.parse(xhr.response);
               if (obj.error) throw 'error';
                obj.meta = { repo: repo, path: '/' };
               // browseResult.js#zoektBrowseResultRender
               zoektBrowseResultRender(div, obj);
            } catch (e) {
               div.innerHTML = '(internal error)';
            }
         });
      }
   });
}

function utilBuildSize(size) {
   var lv = ['B', 'KB', 'MB', 'GB', 'TB'];
   var index = 0;
   while (size >= 1024 && index < lv.length - 1) {
      size /= 1024;
      index ++;
   }
   size = (~~(size * 1000)) / 1000;
   return size + lv[index];
}

zoektSearchEvents();

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

var zoektSpace = /\s/;
var zoektStop = /[`~!@#$%\^&*()-+=|\\{}\[\]<>:;"',.\/?]/;
function zoektFileRender(elem, text) {
   var div, lastSpan;
   text.split('\n').forEach(function (line) {
      div = document.createElement('div');
      div.appendChild(document.createTextNode(line));
      /* - tokenize source code
      // 0 = lastCh, 1 = lastSpace, 2 = lastStop
      var stat = 0;
      lastSpan = null;
      line.split('').forEach(function (ch) {
         if (zoektSpace.test(ch)) {
            if (stat !== 1) {
               lastSpan = document.createElement('span');
               div.appendChild(lastSpan);
            }
            lastSpan.appendChild(document.createTextNode(ch));
            stat = 1;
         } else if (zoektStop.test(ch)) {
            lastSpan = document.createElement('span');
            div.appendChild(lastSpan);
            lastSpan.appendChild(document.createTextNode(ch));
            stat = 2;
         } else {
            if (stat !== 0 || !lastSpan) {
               lastSpan = document.createElement('span');
               div.appendChild(lastSpan);
            }
            lastSpan.appendChild(document.createTextNode(ch));
            stat = 0;
         }
      });
      */
      if (!line.length) div.innerHTML = '&nbsp;';
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

</script>
</body>
</html>
