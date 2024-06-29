<?xml version="1.0" encoding="ISO-8859-1"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output method="html"/>
  <xsl:template match="/testsuites">
    <html>
      <head>
        <title>Test Report</title>
      </head>
      <body>
        <h2>Test Report</h2>
        <table border="1">
          <tr>
            <th>Test Suite</th>
            <th>Test Case</th>
            <th>Time (s)</th>
            <th>Result</th>
          </tr>
          <xsl:for-each select="testsuite/testcase">
            <tr>
              <td><xsl:value-of select="../../@name"/></td>
              <td><xsl:value-of select="@name"/></td>
              <td><xsl:value-of select="@time"/></td>
              <td>
                <xsl:choose>
                  <xsl:when test="failure">Failed</xsl:when>
                  <xsl:otherwise>Passed</xsl:otherwise>
                </xsl:choose>
              </td>
            </tr>
          </xsl:for-each>
        </table>
      </body>
    </html>
  </xsl:template>
</xsl:stylesheet>
