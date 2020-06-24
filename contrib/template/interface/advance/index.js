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
