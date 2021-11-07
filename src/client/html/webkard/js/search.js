var searchjs = {
    data: {
        scope: $("#searchPage"),
        queryString: basejs.getQueryString(),
        loadMorePars: {
            offOn: false,
            scrollType:"",
            scrollId:""
        }
    },
    init: function () {
        var _this = this;
        _this.bindIndexResult();
        _this.bindSearchResult();
        _this.bindScrollResult();

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
    getHttpPars:function(url,data,scrollType,clearHtml,$loadMore){
        var _this = this;
        
        var httpPars = {
            url: url,
            data:data,
            success: function (resultDto) {
        
                if (!resultDto||resultDto.code!=200) {
                    return;
                }

                var $htmlObj=$(".search-result-left", _this.data.scope);

                if(resultDto.data==null){
                    if(clearHtml){
                        $htmlObj.html("");
                        $loadMore.text("查无数据");
                    }else{
                        _this.data.loadMorePars.offOn = false;
                        $loadMore.text("已经是底部");
                    }
                    return;
                }

          
                var resultHtml= _this.getResultHtml(clearHtml,resultDto.data.search_hits);
                if (clearHtml) {
                    $htmlObj.html(resultHtml);
                } else {
                    $htmlObj.append(resultHtml);
                }

                _this.data.loadMorePars.scrollId = resultDto.data.scroll_id;
                _this.data.loadMorePars.scrollType=scrollType;
                _this.data.loadMorePars.offOn = true;
                $loadMore.text("加载更多");
                
            },
            error: function () {
                _this.data.loadMorePars.offOn = false;
                $(".search-result-left", _this.data.scope).empty();
            }
        };
        return httpPars;
    },
    bindIndexResult: function () {
     
        var _this = this;

        var $loadMore = $(".load-more>span", _this.data.scope);
        $loadMore.text("加载中...");
 
        var url=basejs.requestDomain + "/home/index";
        var data={page_count:10};
        var indexHttpPars=_this.getHttpPars(url,data,"index",true,$loadMore);

        var indexHttpHelper = new httpHelper(indexHttpPars);
        indexHttpHelper.send();

    },
    bindSearchResult: function () {
        var _this = this;

        var $loadMore = $(".load-more>span", _this.data.scope);

        var searchHttpPars=_this.getHttpPars(basejs.requestDomain + "/home/search",{},"search",true,$loadMore);
        var indexHttpPars=_this.getHttpPars(basejs.requestDomain + "/home/index",{page_count:10},"index",true,$loadMore);

        $(".btn-search", _this.data.scope).click(function () {
           debugger;
            _this.data.loadMorePars.offOn = false;

            var httpPars={};
            var keyword=$("#searchBox", _this.data.scope).val()
            if(keyword&&keyword.length>0){
                searchHttpPars.data={
                    search_word:keyword,
                    page_count:10
                };
                httpPars=searchHttpPars;
            }else{
                httpPars=indexHttpPars
            }

            $loadMore.text("加载中...");
            var h = new httpHelper(httpPars);
            h.send();
        });

        $("body").keydown(function (e) {
            if (e.keyCode == "13") {//keyCode=13是回车键；数字不同代表监听的按键不同
                $(".btn-search", _this.data.scope).click();
            }
        });

    },
    bindScrollResult:function(){
        var _this = this;
        var $loadMore = $(".load-more>span", _this.data.scope);

        var scrollIndexHttpPars=_this.getHttpPars(basejs.requestDomain + "/home/scroll_index",{},"index",false,$loadMore);
        var scrollSearchHttpPars=_this.getHttpPars(basejs.requestDomain + "/home/scroll_search",{},"search",false,$loadMore);

        $loadMore.loadMore(10, function () {
            debugger;
            //这里用 [ off_on ] 来控制是否加载 （这样就解决了 当上页的条件满足时，一下子加载多次的问题啦）
            if (_this.data.loadMorePars.offOn) {
                _this.data.loadMorePars.offOn = false;

                var httpPars={};
                switch(_this.data.loadMorePars.scrollType){
                    case "index":  httpPars=scrollIndexHttpPars;break;
                    case "search": httpPars=scrollSearchHttpPars;break;
                }
                httpPars.data={"scroll_id": _this.data.loadMorePars.scrollId }

                $loadMore.text("加载中...");
                var h = new httpHelper(httpPars);
                h.send();
            }
        });
    },
    getResultHtml: function (clearHtml,searchHitDtos) {

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
                    texts: texts,
                    creationTime: basejs.getDateDiff(basejs.getDateTimeStamp(searchHitDto.creation_time))
                });

                resultHtml += resultRowHtml;
                resultRowHtml = "";

            }

            
        }

        return resultHtml;
    } 
    
};

$(function () {
    //菜单
    topMenu.bindMenu();

    searchjs.init();
});