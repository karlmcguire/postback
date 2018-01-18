<?php
// config.php contains constants REDIS_ADDRESS, REDIS_PORT, and REDIS_AUTH (I'm
// not including them in this repo for security reasons)
include "config.php";

// verify that the request is POST
if($_SERVER["REQUEST_METHOD"] != "POST") {
	// method not allowed
	http_response_code(405);
	header("Allow: POST");
	return;
}

// verify that the Content-Type is application/json
if($_SERVER["CONTENT_TYPE"] != "application/json") {
	// unsupported media type
	http_response_code(415);
	return;
}

// attempt to decode the json request from the body
$request = json_decode(file_get_contents("php://input"), true);

// verify that it is valid json
if(!isset($request)) {
	// bad request
	http_response_code(400);
	return;
}

// check that all the data we need is there
if(!isset($request["endpoint"]["method"]) ||
	!isset($request["endpoint"]["url"])   ||
	!isset($request["data"])) {
	// bad request
	http_response_code(400);
	return;
}


?>