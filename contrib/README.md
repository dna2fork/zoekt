# contrib
==========

### how to avoid ctags error when parsing symbols

add .ctags.d/avoid-error.ctags

```
--languages=-TOML
--languages=-Cargo
```

### migrate to sourcegraph version of zoekt

> ref: https://github.com/sourcegraph/zoekt

- merge commits into 3 commit
   1. support multiple repo selector e.g. r:project1,project2,...
   2. add contrib for customized APIs and UI
   3. create HTTP entries to bind new APIs for webserver

```
commit 22ef1ec2bfe6ae37573d0602e92d9f86a0556e1e
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jan 12 16:48:21 2021 +0800

    bugfix: occurrence report, avoid run many at once

commit 819dfbd58911bfe0b4a1310c2d9e5dd27f43f054
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jan 12 14:27:18 2021 +0800

    bugfix: occurrence report, _ is not a stop

commit 304a514b262db29952b373b9b4cc7bc79349eec3
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Jan 7 14:30:05 2021 +0800

    fork zoekt-webserver to contrib/cmd/zoekt-webserver

commit 9da4a93935da53de543153ee59f2d185865787ab
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Jan 7 14:29:52 2021 +0800

    bugfix: polish occurrence report
    
    if report not exist, should show report status:
       generating or not

commit f5bb54edf754a737ad19e288952a919a0b245dc9
Author: Seven Lju <dna2gle@gmail.com>
Date:   Mon Jan 4 14:56:30 2021 +0800

    feature: add occurrence report
    
    - get occurrence report
    - simply scan source code to match specified words

commit 1fada244cf1fd7df2368dcaf98c7fc9864943119
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Oct 22 03:04:16 2020 -0400

    support ctags scope
    
    - in many langauges, there is hierarchy for symbols
      for example in Java, we can define methods in a class
      thus, we need to support ctags scope to track this kind of info

commit 16c0574441c3bdd74513052d60ed3847271ca16a
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Oct 22 02:35:22 2020 -0400

    bugfix: incorrect interface nil check
    
    - in web/contrib_contents.go
      - `project == nil`, where project is an interface var `analysis.IProject`
      - use switch to cast to real type and check nil

commit 315fac42e91eb6830747b5fc37c72760a17a1625
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Oct 22 01:48:43 2020 -0400

    feature: restful api to dump ctags output
    
    - list all symbols in a source file via ctags

commit 775381795559ecb19f6146fc1370ce9bc4193a4a
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Sep 1 17:26:19 2020 +0800

    add project list api
    
    - get project list (not indexed project list)

commit cdc9731e2b381758b884887b57894d1ff139646f
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Sep 1 17:01:32 2020 +0800

    add reindex api
    
    - to make reindex possible for remote management
    - a=reindex, r=<project>

commit b7f5606ff121e9770f4c163acbf50d9e0e021b14
Author: Seven Lju <dna2gle@gmail.com>
Date:   Mon Aug 31 15:44:16 2020 +0800

    feature: add track api
    
    - track point CRUD
    - a=track r=<project> f=<path> l=<line>
      - c=get l<=0 -> read track point list for the specified file
      - c=get l>0  -> read track point at the specific line for the specified file
      - c=sync     -> check track point metadata are up-to-date
      - c=add      -> add a track point at the specified line
      - c=del      -> del a track point at the specified line

commit 1944871ece50ebb50dc1d752ca0d52c0ad5536bd
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Aug 27 15:20:27 2020 +0800

    feature: init track line
    
    - add lib functions for track moved line
    - guess best line if map to multiple lines
    - return -1 for line deleted or non matched

commit f7e14f52411a42068693291ed472717003bc11e6
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Aug 26 18:16:04 2020 +0800

    build: one script to build zoekt index and webserver binaries

commit ed6fdf70f11d383fe54af5dce626fb27f78f70e7
Author: Seven Lju <dna2gle@gmail.com>
Date:   Mon Aug 17 01:39:26 2020 -0400

    feature: support timestamp in blame info
    
    - parse timestamp info for blame
    - add a field for BlameDetails

commit e7245b6662c3cb2937140627b5db35db590f3467
Author: Seven Lju <dna2gle@gmail.com>
Date:   Fri Aug 14 05:39:24 2020 -0400

    bugfix: git blame may have file name info
    
    - if git blame have file rename info
      all author and commit will be unkown

commit a0abe5f9065dc39650163148b5dee809d997d735
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jul 21 12:47:23 2020 -0400

    p4 project: cache p4 client to local disk
    
    - cache p4 client details in local
      - if has cache, read from local instead of reading from remote server via p4 command
      - this is to acclerate getDetails (no connection to p4 server required afterwords)

commit a193788557e8a30279e20c1809a60676722d9f21
Author: Seven Lju <dna2gle@gmail.com>
Date:   Mon Jun 29 00:34:07 2020 -0400

    minior: fixed bug for map view to append tail / if needed

commit 446f9121bff2a3b7392526e5784946c858160c0d
Author: Seven Lju <dna2gle@gmail.com>
Date:   Sat Jun 27 01:42:01 2020 -0400

    ui minior: support back to repo list in a repo root
    
    - click .. to back from a repo root to repo list
    - set tab size to 4 in source view

commit 0e61ef2f16a6a4833d778b6a0268e1157de56cf0
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 24 09:56:05 2020 -0400

    push advance ui to search.html.tpl

commit 8a71147920326df061b663f3e86e22c675ea13d6
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 24 09:44:11 2020 -0400

    imporve advance ui for search and browse

commit 60937873a1ad3100970ba1ace7a2732cee18023f
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 24 08:51:09 2020 -0400

    support search view in advance ui

commit 23aa105a3856361c663d023c195e33d87f7baa89
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 24 07:40:40 2020 -0400

    bugfix: wrong up dir for file and wrong format for git dir list

commit dedc174974ab936d920dab0dee3e8fe44de2d034
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 24 06:47:28 2020 -0400

    support advance ui for zoekt

commit 370890b249f6ee92a65302e77ad3c0398efd2755
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 24 04:58:02 2020 -0400

    split search.html.tpl into parts and use script to compile it
    
    - basic ui
      - search: search box + search result

commit eb4e3d6c16f34e34893fc238f959b5ffe85f0ac3
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 24 01:34:07 2020 -0400

    add index function in contrib
    
    - copy and modify cmd/zoekt-index/main.go into a function in contrib/util.go

commit 777ed6819be9f3815fff690cf12283eb232e1d40
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 24 00:51:58 2020 -0400

    move util to contrib from contrib/analysis

commit 048dc95ed54a8e800ceee2c75dd85ce67b202e6a
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jun 23 22:16:14 2020 -0400

    minor improvement: add branch info for scm server details

commit 0374d6727524c9c644d0396440c7a961285abc17
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jun 23 13:17:28 2020 -0400

    support keyval full text search
    
    - refact search in analysis
       - expose search function to external so that no need dup function
    - refact search renderer in web/server
       - share the same renderer in different search functions
    - add search http handler for keyval
       - remove original http handler in keyval
       - move keyval api router in web/contrib_keyval.go
       - add search api

commit cc5141e5b83807b76490a7c74474544c787b7382
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jun 23 12:01:07 2020 -0400

    add keyval service for zoekt to store source metadata
    
    - copy keyval from
      - https://github.com/dna2zodiac/keyval @ 70a4e217b7cc953624131503061b171d2ffd375a

commit 1f6047405eef9dab779d209b59d95563dcedf9d9
Author: Seven Lju <dna2gle@gmail.com>
Date:   Fri Jun 19 01:57:41 2020 -0400

    support get remote link for a specific project + path
    
    - p4: project + path -> server path
    - git: project -> remote url + path

commit 00a50ef761b1610e7207b26b99b12b325fe8a5fd
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jun 2 18:12:33 2020 +0800

    [core] matchtree to support word match for sym
    
    - `sym:^create$` will only match definitions matching whole word
      - include: create() {}
      - exclude: createObject() {}
    - use this until below resolved
      - https://gerrit-review.googlesource.com/c/zoekt/+/233092

commit f7e1cc6059b08e27a8b37848f7247a3ecab66feb
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 17 12:48:24 2020 -0400

    bugfix: silly bug for invalid json in scm send dir contents

commit d9ddfe401da340da13b34395ec3d452ba63d8dd2
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jun 16 12:20:55 2020 -0400

    support commit search
    
    - commit search only
      - /path/to/repo/.<type>/.zoekt/index/commit...
        e.g. /data/zoekt/.git/.zoekt/index/commit_v15.00000.zoekt
    - add search action to web server

commit 903ea14744b9db8bb9f7f45da09dfb9286be3528
Author: Seven Lju <dna2gle@gmail.com>
Date:   Sun Jun 14 16:41:59 2020 +0800

    [analysis] bugfix: git status -s

commit fd8d865b4715641ef050b73ac251b481d7da1787
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Jun 4 22:58:43 2020 +0800

    [analysis][core/server] support blame/commit history
    
    - support api for blame and commit history info
    - bugfix for blame, commit

commit 440875626a29ac7aa433ba3b390336f48867ac6d
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Jun 4 15:58:24 2020 +0800

    [analysis] support get p4/git commit details

commit 3a070a1820b1d77733903ec88cd1c717843cfec0
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Jun 4 12:46:39 2020 +0800

    [server contents] fix silly error: missing return after return error

commit 4d51e9e2432d4c131f1bf246365b9e0aca3f86f1
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Jun 4 12:01:05 2020 +0800

    Revert "[core] matchtree to support word match for sym"
    
    This reverts commit 6d21d8d99efdcde436a6e76ef087e5d06292dc59.
    - regexp will be supported in `sym` filter
    - ref: https://gerrit-review.googlesource.com/c/zoekt/+/233092/3/

commit 6ea680460e12bc150516a5dd3bf86da9a330937a
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 3 22:48:08 2020 +0800

    [analysis] if git dir exists, initialize branch if not provided

commit b8145c4fb2ccfb211496c4fc25bd4fe18159d94d
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 3 22:46:51 2020 +0800

    [core/server] support get contents at specific rev
    
    - get folder with rev
    - get file with rev

commit e479336c672efe9ebb44c52196f76d186dff5c6e
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 3 17:01:21 2020 +0800

    [core/web] move changes to standalone file
    
    - move fsprint api to standalone file of contrib_contents.go
    - move basic auth code to contrib_contents.go
    - init scmprint api
      - added action, revision

commit dfb633b7c8d5359e4649acc3ca5a9b39a7d8bea8
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jun 3 16:59:59 2020 +0800

    [analysis] list files in directory at specific rev

commit 6d21d8d99efdcde436a6e76ef087e5d06292dc59
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jun 2 18:12:33 2020 +0800

    [core] matchtree to support word match for sym
    
    - `sym:^create$` will only match definitions matching whole word
      - include: create() {}
      - exclude: createObject() {}

commit f1f02e67776ad8d5064f4594f38abaddd8a73817
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jun 2 00:01:51 2020 +0800

    [analysis] support generic new project
    
    - from different project type
      - git
      - p4

commit 07f559cc412d02f8b356019ea22b662c6de3de3c
Author: Seven Lju <dna2gle@gmail.com>
Date:   Sun May 31 23:00:35 2020 +0800

    [analysis] support file hash, len, contents, ...
    
    - support get file in p4/git
      - hash at specific rev
      - len at rev
      - contents at rev
      - blame from line to another line at rev
      - commit

commit 45948f4fdb0818b2f9d4e27d97907764c84fb2f3
Author: Seven Lju <dna2gle@gmail.com>
Date:   Sun May 31 13:08:36 2020 +0800

    [analysis] repo sync will return what file modified

commit 0d52f09b9ea42340c09d1be517d42491aecd9797
Author: Seven Lju <dna2gle@gmail.com>
Date:   Sun May 31 11:42:10 2020 +0800

    [analysis] add file hash and length calc func

commit f0cc31b1dd662809ef5019ba651f98a65d906123
Author: Seven Lju <dna2gle@gmail.com>
Date:   Sat May 30 17:03:17 2020 +0800

    bugfix: correct isBinary check
    
    - use []rune instead of []byte to check '\x00'

commit aadfeefd336c7ae7787b7e4a3164581a1ecceff2
Author: Seven Lju <dna2gle@gmail.com>
Date:   Sat May 30 16:33:37 2020 +0800

    [analysis] add IsBinaryFile check
    
    - refact exec functions
    - add check binary file func

commit f0825d375429bf9c932e371d2b308876ad6a0ddc
Author: Seven Lju <dna2gle@gmail.com>
Date:   Fri May 29 21:51:22 2020 +0800

    [analysis] init analysis package
    
    - support p4/git project sync

commit d8ae9be38eb6d652257664fd07dfb80b4863419b
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Apr 7 19:29:35 2020 +0800

    [server] follow soft link to check directory

commit ddfb136b19238e3e5c3dcba599ee32b6aaa178e2
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Mar 31 23:33:05 2020 +0800

    [parser] make sure we can search from root '/'
    
    - e.g. `r:test f:/lv1/lv2 test`

commit 3c627abbb1c2049d0308ec35d2fdcc1b2f28057f
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Mar 25 23:57:53 2020 +0800

    [server] improve user experience
    
    tranform query for better user experience:
    - ( <-> \(
    - ) <-> \)
    - [ <-> \[
    - ] <-> \]

commit 75f938ee4753123d642229f50d3845f3a248ee74
Author: Seven Lju <dna2gle@gmail.com>
Date:   Mon Mar 9 00:02:11 2020 +0800

    [template] add search and about template
    
    - add search template
    - support search with basic auth
    - add about template

commit 126e0509902f9278f79987b1fb4ebe0449fecc00
Author: Seven Lju <dna2gle@gmail.com>
Date:   Sun Mar 8 00:25:36 2020 +0800

    [webserver] add basic auth

commit 8dde201573c2c1d88763b8dc270227c15af98072
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Mar 5 14:28:03 2020 +0800

    [webserver] combine one item directory
    
    - if a directory contains only one directory item, combine them
      e.g. main/java/utils/{...}
      main/ -> main/java/ -> main/java/utils

commit d129f42d940c7186f062d7be589b7d35e02fefe0
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jan 15 16:42:13 2020 +0800

    bugfix: make sure JsonText filter works to output correct json data

commit b854817062a746cceb3718b614c00a96c3a13c1d
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jan 15 16:18:49 2020 +0800

    [template] high light matched text

commit ab504a3b2c4aab6019567a7e2d5f25d2828e8727
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jan 8 15:27:12 2020 +0800

    bugfix: typo \\t instead of \\r and \" -> \\\"

commit 19b1efb505ef13ee9f37db8b5a90556cc38d2b0a
Author: Seven Lju <dna2gle@gmail.com>
Date:   Wed Jan 8 12:41:27 2020 +0800

    improve jsonText formatter

commit af3f5d44e921a1a2e372b7b2c8886885c4878a3e
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jan 7 15:27:04 2020 +0800

    security fix: validate path to prevent contents dump

commit 0b960f0b236ae718d31df1ad95bc7c34bf044d1d
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Jan 7 15:15:55 2020 +0800

    separate jsonText for template and web server usage

commit 4dce51c916d972dd6e6ebd64864caee9a6c110ab
Author: Seven Lju <dna2gle@gmail.com>
Date:   Mon Jan 6 23:34:57 2020 +0800

    add template and implement file server
    
    - add template for repo list and file contents
    - add `/fsprint` api to fetch file and directory contents

commit 315dad835280d1305b24a7dd792f8f4d87b56647
Author: Seven Lju <dna2gle@gmail.com>
Date:   Mon Jan 6 15:17:17 2020 +0800

    support search in multiple repos
    
    - "r:a,b,c"; comma separate list
    - multiple "r" may cause unexpected results "r:a r:b r:c ..."

commit 089d4f1931aa8eb346dfa5b3595189cb1fb2e739
Author: Seven Lju <dna2gle@gmail.com>
Date:   Thu Oct 17 10:16:48 2019 +0800

    add JsonText to support json template

commit 2b3bc4ed0bd0544d43f889d69ab793899bbbaa7b
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Apr 30 11:13:25 2019 +0800

    move to a better folder to build

commit abb2e6e5c89f2492c8658fc1db938ca61669a19a
Author: Seven Lju <dna2gle@gmail.com>
Date:   Tue Apr 30 11:08:43 2019 +0800

    presist dependencies
```

(old backup: branch google-master)
