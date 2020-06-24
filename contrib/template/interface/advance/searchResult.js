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
   pre.appendChild(document.createTextNode(JSON.stringify(obj, null, 3)));
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
