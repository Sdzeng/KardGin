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
            "<div class='introduce-info'><div class='introduce-info-pic'></div></div>" +
            "<div class='result-entity'>" +
            "<div class='result-info'>" +
            "<div class='result-title'><a href='#{detailPage}'>#{title}</a></div>" +
            "<div class='result-subtitle'><a href='#{detailPage}'>又名：#{subtitle}</a></div>" +
            "<div class='result-tab'><span class='subtitls-startat'>下面片段摘自 #{startAt}</span><span class='subtitls-lan'>#{lan}</span></div>" +
            "<div class='result-content'><a href='#{detailPage}' class='essay-content'>#{texts}</a></div>" +
            "<div class='result-footer'>#{creationTime}更新</div>" +
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

                var $htmlObj=$(".search-result-warp-list", _this.data.scope);

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

          
                var resultHtml= _this.getResultHtml(resultDto.data.search_hits);
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
                $(".search-result-warp-list", _this.data.scope).empty();
            }
        };
        return httpPars;
    },
    bindIndexResult: function () {
     
        var _this = this;

        var $loadMore = $(".load-more>div", _this.data.scope);
        $loadMore.text("加载中...");
 
        var url=basejs.requestDomain + "/home/index";
        var data={page_count:5};
        var indexHttpPars=_this.getHttpPars(url,data,"index",true,$loadMore);

        var indexHttpHelper = new httpHelper(indexHttpPars);
        indexHttpHelper.send();

    },
    bindSearchResult: function () {
        var _this = this;

        var $loadMore = $(".load-more>div", _this.data.scope);

        var searchHttpPars=_this.getHttpPars(basejs.requestDomain + "/home/search",{},"search",true,$loadMore);
        var indexHttpPars=_this.getHttpPars(basejs.requestDomain + "/home/index",{page_count:5},"index",true,$loadMore);

        $(".btn-search", _this.data.scope).click(function () {
            _this.data.loadMorePars.offOn = false;

            var httpPars={};
            var keyword=$("#searchBox", _this.data.scope).val()
            if(keyword&&keyword.length>0){
                searchHttpPars.data={
                    search_word:keyword,
                    page_count:5
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
        var $loadMore = $(".load-more>div", _this.data.scope);

        var scrollIndexHttpPars=_this.getHttpPars(basejs.requestDomain + "/home/scroll_index",{},"index",false,$loadMore);
        var scrollSearchHttpPars=_this.getHttpPars(basejs.requestDomain + "/home/scroll_search",{},"search",false,$loadMore);

        $loadMore.loadMore(2, function () {
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
    getResultHtml: function (searchHitDtos) {
        var _this = this;
        var resultHtml = "";

        if (searchHitDtos) {
            var resultRowHtml = "";

            for (var index in searchHitDtos) {
                var searchHitDto = searchHitDtos[index];
                var detailPage = "/detail.html?path_id=" + searchHitDto.path_id;
         
                var texts="";
               if(searchHitDto.texts&&searchHitDto.texts.length>0) {
                    texts="【"+searchHitDto.texts.join("】【")+"】"
                }

                var pick="";
                switch(index%1024){
                    case 0:pick="🍑🍓🥝";break;
                    case 1:pick="🎄🎃";break;
                    case 2:pick="🍕";break;
                    // case 0:pick="🍑🍓🥝";break;
                    // case 1:pick="🎅🎄🎃";break;
                    // case 2:pick="🍕🧁🍵";break;
                    // case 3:pick="🍉";break;
                    // case 4:pick="🎅";break;
                    // case 5:pick="🥝";break;
                    // case 6:pick="🎄";break;
                    // case 7:pick="🎃";break;
                }
                var creationTime=pick+" "+basejs.getDateDiff(basejs.getDateTimeStamp(searchHitDto.create_time));

                resultRowHtml += _this.template.searchResultRow.format({
                    startAt: basejs.formatSeconds(searchHitDto.start_at),
                    lan: searchHitDto.lan,
                    detailPage: detailPage,
                    title: searchHitDto.title,
                    subtitle: searchHitDto.subtitle,
                    texts: texts,
                    creationTime: creationTime
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