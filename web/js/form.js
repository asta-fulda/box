// Generated by CoffeeScript 1.4.0

/*                                                                                                                                                                                                                 
 * Copyright 2011 Dustin Frisch<fooker@lab.sh>
 * 
 * This file is part of box.
 * 
 * box is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * 
 * box is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * 
 * You should have received a copy of the GNU General Public License
 * along with box. If not, see <http://www.gnu.org/licenses/>.
*/


(function() {
  var AnswerModel, ErrorModel, Model, TrackingModel;

  ko.bindingHandlers.readonly = {
    update: function(element, valueAccessor) {
      if (ko.utils.unwrapObservable(valueAccessor())) {
        return element.setAttribute("readonly", true);
      } else {
        return element.removeAttribute("readonly");
      }
    }
  };

  TrackingModel = (function() {

    function TrackingModel(id) {
      var _this = this;
      this.id = id;
      this.size = ko.observable(0);
      this.received = ko.observable(0);
      this.state = ko.observable('starting');
      this.progress_url = '/progress?X-Progress-ID=' + this.id;
      this.interval = setInterval((function() {
        return $.ajax({
          'url': _this.progress_url,
          'dataType': 'json',
          'success': function(data) {
            _this.size(data != null ? data.size : void 0);
            _this.received(data != null ? data.received : void 0);
            return _this.state(data != null ? data.state : void 0);
          }
        });
      }), 1000);
      ko.computed(function() {
        if (_this.state() === 'done' || _this.state === 'error') {
          return clearInterval(_this.interval);
        }
      });
      this.progress = ko.computed(function() {
        switch (_this.state()) {
          case 'starting':
            return '0%';
          case 'done':
            return '100%';
          case 'error':
            return '0%';
          case 'uploading':
            return "" + (Math.round(_this.received() / _this.size() * 100.0 * 100.0) / 100.0) + "%";
          default:
            return '0%';
        }
      });
      this.state_text = ko.computed(function() {
        switch (_this.state()) {
          case 'starting':
            return 'Warten...';
          case 'done':
            return 'Fertig';
          case 'error':
            return 'Fehler!';
          case 'uploading':
            return _this.progress();
          default:
            return '...';
        }
      });
    }

    return TrackingModel;

  })();

  AnswerModel = (function() {

    function AnswerModel(data) {
      this.id = data.upload_id;
      this.user = data.upload_user;
      this.file = data.upload_file;
      this.size = data.upload_size;
      this.expiration = data.upload_expiration;
      this.url = "https://box.hs-fulda.org/download/" + this.id + "?dl=" + this.file;
    }

    return AnswerModel;

  })();

  ErrorModel = (function() {

    function ErrorModel(data) {
      this.code = data.error_code;
    }

    return ErrorModel;

  })();

  Model = (function() {

    function Model() {
      var i,
        _this = this;
      this.upload_tracking_id = ((function() {
        var _i, _results;
        _results = [];
        for (i = _i = 1; _i <= 32; i = ++_i) {
          _results.push(Math.floor(Math.random() * 16).toString(16));
        }
        return _results;
      })()).reduce(function(t, s) {
        return t + s;
      });
      this.target_url = '/upload?X-Progress-ID=' + this.upload_tracking_id;
      this.file = ko.observable('');
      this.title = ko.observable('');
      this.description = ko.observable('');
      this.username = ko.observable('');
      this.password = ko.observable('');
      this.terms_accepted = ko.observable(false);
      this.data_valid = ko.computed(function() {
        return _this.file() !== '' && _this.username() !== '' && _this.password() !== '' && _this.terms_accepted();
      });
      this.tracking = ko.observable(null);
      this.answer = ko.observable(null);
      this.error = ko.observable(null);
    }

    Model.prototype.open_file_chooser = function() {
      return $('#file').click();
    };

    Model.prototype.start_upload = function() {
      this.tracking(new TrackingModel(this.upload_tracking_id));
      return true;
    };

    Model.prototype.upload_completed = function(data, event) {
      var _ref;
      if ((_ref = this.tracking()) != null) {
        _ref.state('done');
      }
      data = JSON.parse(event.target.contentDocument.body.innerText || event.target.contentDocument.body.textContent);
      if (data.error_code) {
        return this.error(new ErrorModel(data));
      } else {
        return this.answer(new AnswerModel(data));
      }
    };

    return Model;

  })();

  $(function() {
    return ko.applyBindings(new Model());
  });

}).call(this);
