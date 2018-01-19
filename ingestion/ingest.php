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

// the count field is added so the delivery agent knows when to stop listening
// for data objects
$request["endpoint"]["count"] = sizeof($request["data"]);

// create a redis object using variables defined in config.php
$redis = new Redis();
$redis->pconnect(REDIS_ADDRESS, REDIS_PORT);
$redis->auth(REDIS_AUTH);

// postback:[uuid]
$postback_key = "postback" . uniqid();
// postback:[uuid]:data
$postback_data_key = $postback_key . ":data";

// watch for collisions (very rare but possible)
$redis->watch($postback_key);

// push postback:[uuid] to postbacks list, notifying delivery agent there's a 
// new postback object to handle
$ret = $redis->multi()
	->set($postback_key, json_encode($request["endpoint"]))
	->rpush("postbacks", $postback_key)
	->exec();

// make sure redis transaction was successful
if($ret[0] == false || $ret[1] == false) {
	throw new Exception("problem adding to redis");
}

// push each data object to postback:[uuid]:data for delivery agent to handle
foreach($request["data"] as $data) {
	$redis->rpush($postback_data_key, json_encode($data));
}
?>