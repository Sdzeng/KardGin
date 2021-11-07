var searchjs = {
    data: {
        scope: $("#searchPage"),
        queryString: basejs.getQueryString(),
        loadMorePars: {
            //设置essays加载更多
            offOn: false,
            page: 1

        }
    },
    init: function () {
        var _this = this;
 
        _this.bindSearchResult();
 

        $('.go-to-top', _this.data.scope).goToTop();
    },
    template: {
        searchResultRow: ("<div class='search-result-warp'>" +
            "<div class='result-score'><div class='essay-score'>#{startAt}</div><div class='essay-score-head-count'>#{lan}</div></div>" +
            "<div class='result-entity'>" +
            "<div class='result-info'>" +
            "<div class='result-header'><a href='#{detailPage}' class='essay-title'>#{title}</a></div>" +
            "<div class='result-header'><a href='#{detailPage}' class='essay-title'>#{subtitle}</a></div>" +
            "<div class='result-content'><a href='#{detailPage}' class='essay-content'>#{texts}</a></div>" +
            "<div class='result-footer'><span class='essay-nickname'><a>#{lan}</a></span> <span class='essay-creationtime'>#{creationTime}</span></div>" +
            "</div>" +
            "</div >" +
            "</div >")
    }, 
    bindSearchResult: function () {
        var _this = this;

        var $loadMore = $(".load-more>span", _this.data.scope);
        $loadMore.text("加载中...");

        var keyword = null;
        if (_this.data.queryString && _this.data.queryString.keyword && _this.data.queryString.keyword.length > 0) {
            keyword = decodeURI(_this.data.queryString.keyword);
            $("#searchBox", _this.data.scope).val(keyword);
        }
        var httpPars = {
            url: basejs.requestDomain + "/home/search",
            data: { search_word: "老", page_count: 10 },
            success: function (resultDto) {
                debugger;
                //设置essays加载更多
                if (resultDto.code!=200) {
                    return;
                }

                _this.bindResultInfo(resultDto.data.search_hits);

                if (resultDto.data.hasNextPage) {
                    _this.data.loadMorePars.offOn = true;
                    _this.data.loadMorePars.page++;
                    $loadMore.text("加载更多");
                }
                else {
                    _this.data.loadMorePars.offOn = false;
                    $loadMore.text("已经是底部");
                }
            },
            error: function () {
                _this.data.loadMorePars.offOn = true;

                $(".search-result-left", _this.data.scope).empty();
            }
        };

        var essaysHttpHelper = new httpHelper(httpPars);
        essaysHttpHelper.send();

        $(".btn-search", _this.data.scope).click(function () {

            _this.data.loadMorePars.offOn = false;
            _this.data.loadMorePars.page = 1;
            httpPars.data.keyword = $("#searchBox", _this.data.scope).val();
            httpPars.data.pageIndex = _this.data.loadMorePars.page;

            $loadMore.text("加载中...");
            essaysHttpHelper = new httpHelper(httpPars);
            essaysHttpHelper.send();
        });

        $("body").keydown(function () {
            if (event.keyCode == "13") {//keyCode=13是回车键；数字不同代表监听的按键不同
                $(".btn-search", _this.data.scope).click();
            }
        });


        $loadMore.loadMore(50, function () {
            //这里用 [ off_on ] 来控制是否加载 （这样就解决了 当上页的条件满足时，一下子加载多次的问题啦）
            if (_this.data.loadMorePars.offOn) {
                _this.data.loadMorePars.offOn = false;

                httpPars.data.keyword = $("#searchBox", _this.data.scope).val();
                httpPars.data.pageIndex = _this.data.loadMorePars.page;
                $loadMore.text("加载中...");
                essaysHttpHelper = new httpHelper(httpPars);
                essaysHttpHelper.send();
            }
        });

    },
    bindResultInfo: function (searchHitDtos) {

        var _this = this;
        var titleTagArr = [];
        var resultHtml = "";

        if (searchHitDtos) {
            var resultRowHtml = "";

            for (var index in searchHitDtos) {
                var searchHitDto = searchHitDtos[index];
                var detailPage = "/detail.html?path_id=" + searchHitDto.path_id;
         
                var texts="";
               if(searchHitDto.texts&&searchHitDto.texts.length>0) {
                    texts=searchHitDto.texts.join(" ")
                }

                resultRowHtml += _this.template.searchResultRow.format({
                    startAt: searchHitDto.start_at,
                    lan: searchHitDto.lan,
                    detailPage: detailPage,

                    title: searchHitDto.title,
                    subtitle: searchHitDto.subtitle,

                    texts: texts ,
         
                    creationTime: basejs.getDateDiff(basejs.getDateTimeStamp(searchHitDto.creation_time))

                });



                resultHtml += resultRowHtml;
                resultRowHtml = "";

            }

            if (_this.data.loadMorePars.page == 1) {
                $(".search-result-left", _this.data.scope).html(resultHtml);
            } else {
                $(".search-result-left", _this.data.scope).append(resultHtml);
            }
        }


    } 
    
};

$(function () {
    //菜单
    topMenu.bindMenu();

    searchjs.init();
});