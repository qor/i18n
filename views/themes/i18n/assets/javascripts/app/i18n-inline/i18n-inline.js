(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define(['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function (jQuery) {

  'use strict';

  var location = window.location;

  var NAMESPACE = 'qor.i18n.inline';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;

  // For Qor Autoheight plugin
  var EVENT_INPUT = 'input';

  function I18nInlineEdit(element, options) {
    this.jQueryelement = jQuery(element);
    this.options = jQuery.extend({}, I18nInlineEdit.DEFAULTS, jQuery.isPlainObject(options) && options);
    this.multiple = false;
    this.init();
  }

  function encodeSearch(data) {
    var params = [];

    if (jQuery.isPlainObject(data)) {
      jQuery.each(data, function (name, value) {
        params.push([name, value].join('='));
      });
    }

    return params.join('&');
  }

  function decodeSearch(search) {
    var data = {};

    if (search) {
      search = search.replace('?', '').split('&');

      jQuery.each(search, function (i, param) {
        param = param.split('=');
        i = param[0];
        data[i] = param[1];
      });
    }

    return data;
  }

  I18nInlineEdit.prototype = {
    contructor: I18nInlineEdit,

    init: function () {
      var jQuerythis = this.jQueryelement;
      this.makeInputEditable();
      this.bind();
    },

    bind: function () {
      this.jQueryelement.
        on(EVENT_CLICK, jQuery.proxy(this.click, this)).
        on(EVENT_CHANGE, jQuery.proxy(this.change, this));
    },

    unbind: function () {
      this.jQueryelement.
        off(EVENT_CLICK, this.click).
        off(EVENT_CHANGE, this.change);
    },

    makeInputEditable : function () {
      jQuery.fn.editable.defaults.mode = 'popup';
      jQuery.fn.editable.defaults.ajaxOptions = { type: 'POST' };
      this.jQueryelement.editable({
        pk: 1,
        params: function (params) {
          params.Value = params.value;
          params.Locale = jQuery(this).data('locale');
          params.Key = jQuery(this).data('key');
          return params;
        },
        url: '/admin/translations'
      });
      this.jQueryelement.on("hidden", function(e, params) {
        $(this).html($(this).text());
      });
    }
  };

  I18nInlineEdit.DEFAULTS = {};

  I18nInlineEdit.plugin = function (options) {
    return this.each(function () {
      var jQuerythis = jQuery(this);
      var data = jQuerythis.data(NAMESPACE);
      var fn;

      if (!data) {
        jQuerythis.data(NAMESPACE, (data = new I18nInlineEdit(this, options)));
      }

      if (typeof options === 'string' && jQuery.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  jQuery(document).ready(function () {
    I18nInlineEdit.plugin.call(jQuery('.qor-i18n-inline'));
  });

});
