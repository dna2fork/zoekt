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
   var _timer = -1, _hold = false;
   var _x, _y, _tx0, _ty0, _target;
   this.triggers = [];
   this.events = {
      onMouseDown: function (evt) {
         if (that.opt.selectorFn && !that.opt.selectorFn(evt.target)) {
            return;
         }
         _x = evt.offsetX;
         _y = evt.offsetY;
         if (_timer > 0) clearTimeout(_timer);
         _timer = setTimeout(function () {
            _hold = true;
            _target = evt.target;
            _timer = -1;
         }, that.opt.duration);
      },
      onMouseUp: function (evt) {
         if (_timer > 0) {
            clearTimeout(_timer);
            return;
         }
         var hold = _hold, target = _target;
         _hold = false;
         _target = null;
         if (evt.target != target) {
            return;
         }
         if (hold && distance(_x, _y, evt.offsetX, evt.offsetY) < 1) {
            that.fire(createHoldOnEvent(evt.target));
         }
      },
      onMouseLeave: function (evt) {
         if (_timer > 0) {
            clearTimeout(_timer);
            return;
         }
         _hold = false;
         _target = null;
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
            _hold = true;
            _target = evt.target;
            _timer = -1;
         }, that.opt.duration);
      },
      onTouchMove: function (evt) {
         if (!_target) return;
         _tx0 = evt.touches[0].clientX;
         _ty0 = evt.touches[0].clientY;
         if (evt.touches.length > 1 || distance(_x, _y, _tx0, _ty0) >= 1) {
            if (_timer > 0) clearTimeout(_timer);
            _hold = false;
            _target = null;
         }
      },
      onTouchEnd: function (evt) {
         if (_timer > 0) {
            clearTimeout(_timer);
            return;
         }
         if (evt.target !== _target) {
            return;
         }
         var hold = _hold, target = _target;
         _hold = false;
         _target = null;
         if (evt.target != target) {
            return;
         }
         if (hold && evt.cancelable) {
            evt.preventDefault();
            that.fire(createHoldOnEvent(evt.target));
         }
      }
   };
   this.self = dom;
   this.self.addEventListener('mousedown', this.events.onMouseDown);
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
