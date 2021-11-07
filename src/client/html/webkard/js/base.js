var basejs = {
    // requestDomain: "http://192.168.2.211:5000",
    // cdnDomain: "http://192.168.2.211:5000",
    requestDomain: "http://localhost:20191",
    defaults: {
        type: "POST",
        async: true,
        contentType: "application/json;charset=utf-8",//"application/x-www-form-urlencoded"
        traditional: false,
        processData: true,
        data: null,
        loading: "off",
        avatarPath: "/image/default-avatar.jpg"
    },
    getPath: function () {
        var hash = window.location.hash, reg = /^#!/;
        if (reg.test(hash)) {

            return hash.replace(reg, '');
        } else {
            return hash;//storage.local.getItem('redirect') || '';
        }
    },

    getQueryString: function () {
        var queryString = {};
        var name, value;
        var str = location.href; //取得整个地址栏
        var num = str.indexOf("?")
        str = str.substr(num + 1); //取得所有参数   stringvar.substr(start [, length ]

        var arr = str.split("&"); //各个参数放到数组里
        for (var i = 0; i < arr.length; i++) {
            num = arr[i].indexOf("=");
            if (num > 0) {
                name = arr[i].substring(0, num);
                value = arr[i].substr(num + 1);
                queryString[name] = value;
            }
        }
        return queryString;
    },

    lazyInof: function (id) {
        $(id).lazyload({ effect: "fadeIn", threshold: 100 });
    },
    worksDefaultPicInfo: function () {
        return this.src = basejs.cdnDomain + "/image/default/default-picture_190x180.png";
    },
    getDateTimeStamp: function (dateStr) {
        return Date.parse(dateStr.replace(/T/g, ' ').replace(/-/gi, "/"));
    },
    getDateDiff: function (dateTimeStamp) {
        var minute = 1000 * 60;
        var hour = minute * 60;
        var day = hour * 24;
        var halfamonth = day * 15;
        var month = day * 30;
        var now = new Date().getTime();
        var diffValue = now - dateTimeStamp;
        if (diffValue < 0) { return; }
        var monthC = diffValue / month;
        var weekC = diffValue / (7 * day);
        var dayC = diffValue / day;
        var hourC = diffValue / hour;
        var minC = diffValue / minute;
        if (monthC >= 1) {
            result = "" + parseInt(monthC) + "月前";
        }
        else if (weekC >= 1) {
            result = "" + parseInt(weekC) + "周前";
        }
        else if (dayC >= 1) {
            result = "" + parseInt(dayC) + "天前";
        }
        else if (hourC >= 1) {
            result = "" + parseInt(hourC) + "小时前";
        }
        else if (minC >= 1) {
            result = "" + parseInt(minC) + "分钟前";
        } else
            result = "刚刚";
        return result;
    },
    getNumberDiff: function (number) {

        if (number < 1000) {
            result = number;
        }
        else if (number < 10000) {
            result = (number / 1000).toFixed(1) + "k";
        }
        else {
            result = (number / 10000).toFixed(1) + "w";
        }

        return result;
    },
    getStarClass: function (prefix, score) {
        var star = "";
        if (score > 9) {
            star = "50";
        }
        else if (score > 8) {
            star = "45";
        }
        else if (score > 7) {
            star = "40";
        }
        else if (score > 6) {
            star = "35";
        }
        else if (score > 5) {
            star = "30";
        }
        else if (score > 4) {
            star = "25";
        }
        else if (score > 3) {
            star = "20";
        }
        else if (score > 2) {
            star = "15";
        }
        else if (score > 1) {
            star = "10";
        }
        else if (score > 0.5) {
            star = "05";
        }
        else {
            star = "00";
        }


        return (prefix + star);

    },
    arrDistinct: function (arr) {
            var i,
            obj = {},
            result = [],
            len = arr.length;
        for (i = 0; i < arr.length; i++) {
            if (!obj[arr[i]]) { //如果能查找到，证明数组元素重复了
                obj[arr[i]] = 1;
                result.push(arr[i]);
            }
        }
        return result;
    }
};

var httpHelper = function httpHelper() {
    this.init.apply(this, arguments);
};


$.extend(httpHelper.prototype, {

    init: function (opts) {

        var _this = this;
        _this.opts = opts;
        _this.opts.type = opts.type || basejs.defaults.type;
        _this.opts.async = (opts.async == null) ? basejs.defaults.async : opts.async;
        _this.opts.contentType = (opts.contentType == null) ? basejs.defaults.contentType : opts.contentType;
        _this.opts.traditional = (opts.traditional == null) ? basejs.defaults.traditional : opts.traditional;
        _this.opts.processData = (opts.processData == null) ? basejs.defaults.processData : opts.processData;
        _this.opts.data = opts.data || basejs.defaults.data;
        _this.opts.loading = opts.loading || basejs.defaults.loading;

        if (!_this.opts.url) {
            console.error("url is empty");
            alert("url is empty");
            return null;
        }
        else { return _this; }
    },


    // 发送数据
    send: function () {
        var _this = this;
        debugger;
        //return $.Deferred(function ($dfd) {
        $.ajax({
            url: _this.opts.url,
            type: _this.opts.type,
            // xhrFields: {
            //     withCredentials: true //配置http跨域请求中携带cookie
            // },
            crossDomain: true,
            async: _this.opts.async,
            data: JSON.stringify(_this.opts.data),
            contentType: _this.opts.contentType,
            traditional: _this.opts.traditional,
            processData: _this.opts.processData,
            success: function (resultDto, textStatus, jqXHR) {//success
                _this.opts.success && _this.opts.success.apply(this, arguments);
            },
            beforeSend: function () {
                // loading();
                //_this.opts.loading !== 'off' && dialog.loading();
                _this.opts.beforeSend && _this.opts.beforeSend.apply(this, arguments);
            },
            complete: function () {
                // _this.opts.loading !== 'off' && dialog.loading.fade();
                _this.opts.complete && _this.opts.complete.apply(this, arguments);
            },

            error: function (jqXHR, textStatus, errorThrown) {

                if (errorThrown) {
                    console.log(errorThrown);
                }
       
            }
        });
        // });


    }

});

//菜单
var topMenu = {
    bindMenu: function (menuObj) {
        var _this = this;
        _this.splashObj = $('#splash');
        _this.quoteObj = _this.splashObj.children('.quote');
        _this.menuList = $('#menuList');


        //菜单
        //$.when($.getJSON(basejs.requestDomain + '/assets/json/menu.json'), $.get(basejs.requestDomain + '/menu/'))
        $.getJSON('/json/menu.json', function (data) {

            _this.setMenu(data.menu || {});
            // _this.initShow();
        });

    },
    setMenu: function (menuObj) {
        var _this = this;
        function addMenu(item, title, wrapper) {

            for (var key in item) {
                for (var i = 0; i < item[key].length; i++) {
                    var m = item[key][i];
                    var child = $(
                        m.sublist ? "<li class='dropdown'><a class='dropdown-toggle' data-toggle='dropdown' href='javascript:;' >"
                            + m.namech
                            + "<span class='caret'></span></a><ul class='dropdown-menu' aria-labelledby='themes'></ul></li>"
                            : "<li><a href='javascript:;' class='"+(encodeURI(m.url)==location.href?"menu-active":"")+"'>" + m.namech + "</a></li>").appendTo(wrapper)
                        .children();
                      

                    m.title = (title ? (title + ' > ') : "") + m.namech;

                    if (m.url) {
                        _this.menuClick(child.eq(0), m);
                    }

                    if (m.sublist) {
                        addMenu(m.sublist, m.title, child.eq(1));
                    }

                }
            }
        }

        addMenu(menuObj, null, _this.menuList.empty());
    },
    // 触发load
    fireLoad: function (url, title, data) {
        window.location.href = url;

    },
    menuClick: function (target, item) {
        var _this = this;
        target.on('click', function (event, data) {

            console.log(item.title + ":" + item.url);
            _this.fireLoad(item.url, item.title, data);
        });

        if (basejs.getPath().indexOf(item.url) > -1) {

            _this.fireLoad(item.url, item.title);
            target.addClass("active");

        }
    }
}







