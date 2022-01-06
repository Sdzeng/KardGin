// Kard Web Gulpfile

//加载插件
var gulp = require('gulp');
    less = require('gulp-less'),//le编译ss
    minifycss = require('gulp-minify-css'),//压缩css
    // concat = require('gulp-concat'),//合并js
    uglify = require('gulp-uglify'),//压缩js
    rename = require('gulp-rename'),//改输出别名
    del = require('del');//删除文件
 

//执行压缩前，先删除文件夹里的内容
gulp.task('clean', function() {
   return del(['../builds'], {force:true});
});

//  //压缩css
// //gulp.task('minifycss', function() {
// //    return gulp.src('css/*.css')      //压缩的文件
// //        .pipe(minifycss())   //执行压缩
// //        .pipe(rename({suffix: '.min'}))   //rename压缩后的文件名
// //        .pipe(gulp.dest('css'));   //输出文件夹
// //});

//编译less并压缩css
gulp.task('lessminifycss', function() {
    return gulp.src('../less/kard.less')      //压缩的文件
        .pipe(less())    //编译
        .pipe(rename({suffix: '.min'}))   //rename压缩后的文件名
        .pipe(minifycss())   //执行压缩
        .pipe(gulp.dest('../builds/css'));   //输出文件夹
});


//压缩js
gulp.task('minifyjs', function() {
    return gulp.src(['../js/*.js','!../js/*.min.js'])//压缩文件
        //.pipe(concat('main.js'))    //合并所有js到main.js
        //.pipe(gulp.dest('js'))    //输出main.js到文件夹
        .pipe(rename({suffix:'.min'}))//起别名保存
        .pipe(uglify())//压缩
        .pipe(gulp.dest('../builds/js'));//输出文件夹
});
 
//复制.min.js文件
gulp.task('copyminjs', function()  {
    return gulp.src('../js/*.min.js')
        .pipe(gulp.dest('../builds/js'));
});



//监听任务 运行语句 gulp watch
gulp.task('watch',function(){
    gulp.watch('js/*.js',['minifyjs']);//监听js变化
    gulp.watch('css/*.less',['lessminifycss']);//监听css变化
});

//默认命令，在cmd中输入gulp后，执行的就是这个命令
gulp.task('default',gulp.series('clean','lessminifycss','minifyjs','copyminjs'));//按顺序执行相应模块