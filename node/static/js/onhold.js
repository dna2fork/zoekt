'use strict';

// on hold: mouse/touch hold on an element

(function () {

function distance(x1, y1, x2, y2) {
   var dx = x1 - x2, dy = y1 - y2;
   return Math.sqrt(dx * dx + dy * dy);
}

function createHoldOnEvent(target) {
   var evt = new Event('holdon');
   target.dispatchEvent(evt);
   return evt;
}

function ElementHoldEvent(dom, opt) {
   // opt:
   //    selectorFn -- check if it is target
   //    duration   -- how long holding to treat as hold-on
   this.opt = opt || {};
   this.opt.duration = this.opt.duration || 500;

   var that = this;
   var _timer = -1;
   var _x, _y, _tx0, _ty0, _target;
   this.triggers = [];
   this.events = {
      onMouseDown: function (evt) {
         if (that.opt.selectorFn && !that.opt.selectorFn(evt.target)) {
            return;
         }
         _x = evt.offsetX;
         _y = evt.offsetY;
         _target = evt.target;
         if (_timer > 0) clearTimeout(_timer);
         _timer = setTimeout(function () {
            that.fire(createHoldOnEvent(evt.target));
            _target = null;
            _timer = -1;
         }, that.opt.duration);
      },
      onMouseMove: function (evt) {
         if (!_target) return;
         _tx0 = evt.offsetX;
         _ty0 = evt.offsetY;
         if (evt.target !== _target || distance(_x, _y, _tx0, _ty0) >= 1) {
            if (_timer > 0) clearTimeout(_timer);
            _target = null;
            _timer = -1;
         }
      },
      onMouseUp: function (evt) {
         if (_timer > 0) {
            clearTimeout(_timer);
            return;
         }
         _target = null;
         _timer = -1;
      },
      onMouseLeave: function (evt) {
         if (_timer > 0) {
            clearTimeout(_timer);
            return;
         }
         _target = null;
         _timer = -1;
      },
      onTouchStart: function (evt) {
         if (evt.touches.length > 1) return;
         if (that.opt.selectorFn && !that.opt.selectorFn(evt.target)) {
            return;
         }
         _x = evt.touches[0].clientX;
         _y = evt.touches[0].clientY;
         if (_timer > 0) clearTimeout(_timer);
         _timer = setTimeout(function () {
            that.fire(createHoldOnEvent(evt.target));
            _target = null;
            _timer = -1;
         }, that.opt.duration);
      },
      onTouchMove: function (evt) {
         if (!_target) return;
         _tx0 = evt.touches[0].clientX;
         _ty0 = evt.touches[0].clientY;
         if (!evt.cancelable || _target !== evt.target || evt.touches.length > 1 || distance(_x, _y, _tx0, _ty0) >= 1) {
            if (_timer > 0) clearTimeout(_timer);
            _target = null;
            _timer = -1;
         }
      },
      onTouchEnd: function (evt) {
         if (_timer > 0) {
            clearTimeout(_timer);
            return;
         }
         _target = null;
         _timer = -1;
      }
   };
   this.self = dom;
   this.self.addEventListener('mousedown', this.events.onMouseDown);
   this.self.addEventListener('mousemove', this.events.onMouseMove);
   this.self.addEventListener('mouseup', this.events.onMouseUp);
   this.self.addEventListener('mouseleave', this.events.onMouseLeave);
   this.self.addEventListener('touchstart', this.events.onTouchStart);
   this.self.addEventListener('touchmove', this.events.onTouchMove);
   this.self.addEventListener('touchend', this.events.onTouchEnd);
}
ElementHoldEvent.prototype = {
   on: function (fn) {
      if (this.triggers.indexOf(fn) >= 0) return;
      this.triggers.push(fn);
   },
   off: function (fn) {
      var i = this.triggers.indexOf(fn);
      if (i < 0) return;
      this.triggers.splice(i, 1);
   },
   fire: function (evt) {
      this.triggers.forEach(function (fn) {
         fn && fn(evt);
      });
   },
   dispose: function () {
      this.self.removeEventListener('mousedown', this.events.onMouseDown);
      this.self.removeEventListener('mousemove', this.events.onMouseMove);
      this.self.removeEventListener('mouseup', this.events.onMouseUp);
      this.self.removeEventListener('mouseleave', this.events.onMouseLeave);
      this.self.removeEventListener('touchstart', this.events.onTouchStart);
      this.self.removeEventListener('touchmove', this.events.onTouchMove);
      this.self.removeEventListener('touchend', this.events.onTouchEnd);
      this.triggers = [];
   }
};

if (!window.Flame) window.Flame = {};
window.Flame.event = Object.assign(window.Flame.event || {}, {
   ElementHoldEvent: ElementHoldEvent
});

})();
