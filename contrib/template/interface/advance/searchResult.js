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
