<br />
<div align="center">
    <p>HOF BACKEND APPLICATION ERROR HANDLER</p>
<p>To use, you need to call the <i>errorHandler.Format</i> function</p>
<b>For example: </b>
<br />
<code>
    errorHandler.Format(errorHandler.DatabaseError, errors.New("unable to get values"))
</code>
<p>This will always return an error as it implements the GOLANG error interface</p>
</div>
<ul>
<li>Call the function</li>
<li>Call the error code</li>
<li>Create a detailed message</li>
</ul>
