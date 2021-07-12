'use strict';
//@include js/api.js

(function () {

var ui = {
   btn: {
      nav: {
         search: dom('#btn-search'),
         browse: dom('#btn-browse'),
         team: dom('#btn-team'),
         settings: dom('#btn-settings')
      }
   }, // btn
   panel: {
      side: dom('#pnl-side'),
      search: dom('#pnl-side-search'),
      browse: dom('#pnl-side-browse'),
      team: dom('#pnl-side-team'),
      settings: dom('#pnl-side-settings')
   }, // panel
   state: {
      show: function (el) { el.style.display = 'block'; },
      hide: function (el) { el.style.display = 'none'; },
      nav: {
         selected: null,
         select: function (el) { el.classList.add('active'); },
         deselect: function (el) { el.classList.remove('active'); }
      }
   }
};

function onHashChange() {
   console.log(location.hash);
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

function initEvent() {
   window.addEventListener('hashchange', onHashChange);
   ui.btn.nav.search.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.browse.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.team.addEventListener('click', onSwitchSidePanel);
   ui.btn.nav.settings.addEventListener('click', onSwitchSidePanel);
}

function initComponent() {
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
